package di

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"runtime"
)

type (
	di struct {
		beanDefinitionMap map[string]definition  // Name:bean定义
		prototypeMap      map[string]interface{} // Name:初始化的bean
		beanMap           map[string]interface{} // Name:bean实例
		loaded            bool
		unsafe            bool
		valueStore        ValueStore
		beanSort          *list.List
	}
)

var (
	ErrBean       = errors.New("error bean")
	ErrDefinition = errors.New("error definition")
	ErrLoaded     = errors.New("di loaded")
)

func New() DI {
	return &di{
		beanDefinitionMap: map[string]definition{},
		prototypeMap:      map[string]interface{}{},
		beanMap:           map[string]interface{}{},
		valueStore:        van.New(),
		beanSort:          list.New(),
	}
}

func (container *di) UnsafeMode(open bool) DI {
	container.unsafe = open
	return container
}

// RegisterBean 注册一个已生成的bean，根据bean类型生成beanName
func (container *di) RegisterBean(bean interface{}) DI {
	return container.RegisterNamedBean("", bean)
}

// RegisterNamedBean 以指定名称注册一个bean
func (container *di) RegisterNamedBean(beanName string, bean interface{}) DI {
	if !IsPtr(bean) {
		panic(fmt.Errorf("%w: bean must be a pointer", ErrBean))
	}
	if beanName == "" {
		prototype := reflect.TypeOf(bean).Elem()
		if tmpBeanName, ok := (reflect.New(prototype).Interface()).(BeanName); ok {
			if name := tmpBeanName.BeanName(); name != "" {
				beanName = name
			} else {
				beanName = GetBeanName(bean)
			}
		} else {
			beanName = GetBeanName(bean)
		}
	}
	if _, exist := container.beanMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
	}
	container.beanMap[beanName] = bean
	// 加入队列
	container.beanSort.PushBack(beanName)
	return container
}

func (container *di) Provide(prototype interface{}) DI {
	container.ProvideNamedBean("", prototype)
	return container
}

func (container *di) parseBeanType(beanType interface{}) (prototype reflect.Type, beanName string) {
	if IsPtr(beanType) {
		prototype = reflect.TypeOf(beanType).Elem()
	} else {
		prototype = reflect.TypeOf(beanType)
	}
	if tmpBeanName, ok := (reflect.New(prototype).Interface()).(BeanName); ok {
		if name := tmpBeanName.BeanName(); name != "" {
			beanName = name
		} else {
			beanName = GetBeanName(beanType)
		}
	} else {
		beanName = GetBeanName(beanType)
	}
	return
}

func (container *di) ProvideNamedBean(beanName string, beanType interface{}) DI {
	if container.loaded {
		panic(ErrLoaded)
	}
	var prototype reflect.Type
	if beanName == "" {
		prototype, beanName = container.parseBeanType(beanType)
	} else {
		prototype, _ = container.parseBeanType(beanType)
	}

	// 检查bean重复
	if _, exist := container.beanMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
	}
	// 检查beanDefinition重复
	if existDefinition, exist := container.beanDefinitionMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already defined by %s", ErrDefinition, beanName, existDefinition.Type.String()))
	} else {
		container.beanDefinitionMap[beanName] = newDefinition(beanName, prototype)
		// 加入队列
		container.beanSort.PushBack(beanName)
	}
	return container
}

func (container *di) GetBean(beanName string) (interface{}, bool) {
	bean, ok := container.beanMap[beanName]
	return bean, ok
}

func (container *di) NewBean(beanType interface{}) (bean interface{}) {
	_, beanName := container.parseBeanType(beanType)
	return container.NewBeanByName(beanName)
}

func (container *di) NewBeanByName(beanName string) (bean interface{}) {
	def, ok := container.beanDefinitionMap[beanName]
	if !ok {
		panic(fmt.Errorf("%w: %s notfound", ErrDefinition, beanName))
	}
	// 反射实例
	prototype := reflect.New(def.Type).Interface()
	// 注入值
	container.wireValue(reflect.ValueOf(prototype).Elem(), def)
	// 触发BeanConstruct
	if construct, ok := prototype.(BeanConstructWithContainer); ok {
		construct.BeanConstruct(container)
	} else if construct, ok := prototype.(BeanConstruct); ok {
		construct.BeanConstruct()
	}
	// 触发注入 bean
	bean = container.processBean(prototype, def)
	// 初始化完成
	if initialized, ok := bean.(InitializedWithContainer); ok {
		initialized.Initialized(container)
	} else if initialized, ok := bean.(Initialized); ok {
		initialized.Initialized()
	}
	// 使用析构函数来完成 bean 的 destroy
	runtime.SetFinalizer(bean, container.destroyBean)
	return
}

func (container *di) Load() {
	if container.loaded {
		panic(ErrLoaded)
	}

	container.loaded = true
	container.initializeBeans()
	container.processBeans()
	container.initialized()

}

func (container *di) Serve(ctx context.Context) {
	if !container.loaded {
		panic(ErrLoaded)
	}
	<-ctx.Done()
	container.destroyBeans()
}

// initializeBeans 初始化bean对象
func (container *di) initializeBeans() {
	// 创建类型的指针对象
	for beanName, def := range container.beanDefinitionMap {
		container.prototypeMap[beanName] = container.instanceBean(def)
	}
	// 根据排序遍历触发BeanConstruct方法
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		if prototype, ok := container.prototypeMap[beanName]; ok {
			container.constructBean(prototype)
		}
	}
}

// processBeans 注入依赖
func (container *di) processBeans() {
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		if prototype, ok := container.prototypeMap[beanName]; ok {
			def := container.beanDefinitionMap[beanName]
			// 加载为bean
			container.beanMap[beanName] = container.processBean(prototype, def)
		}
	}
}

// initialized 容器初始化完成
func (container *di) initialized() {
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		bean := container.beanMap[beanName]
		container.initializedBean(bean)
	}
}

func (container *di) destroyBeans() {
	// 倒序销毁bean
	for e := container.beanSort.Back(); e != nil; e = e.Prev() {
		beanName := e.Value.(string)
		if bean, ok := container.beanMap[beanName]; ok {
			container.destroyBean(bean)
			delete(container.beanMap, beanName)
		}
		container.destroyBean(e.Value.(string))
	}
}
