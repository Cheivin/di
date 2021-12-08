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
		log               Log
		beanDefinitionMap map[string]definition  // Name:bean定义
		prototypeMap      map[string]interface{} // Name:初始化的bean
		beanMap           map[string]interface{} // Name:bean实例
		loaded            bool
		unsafe            bool
		valueStore        ValueStore
		beanSort          *list.List
		ctx               context.Context
	}
)

func (container *di) Context() context.Context {
	return container.ctx
}

var (
	ErrBean       = errors.New("error bean")
	ErrDefinition = errors.New("error definition")
	ErrLoaded     = errors.New("di loaded")
)

func New() *di {
	return &di{
		log:               stdLogger(),
		beanDefinitionMap: map[string]definition{},
		prototypeMap:      map[string]interface{}{},
		beanMap:           map[string]interface{}{},
		valueStore:        van.New(),
		beanSort:          list.New(),
		ctx:               context.Background(),
	}
}

func (container *di) UnsafeMode(open bool) DI {
	container.unsafe = open
	container.log.Warn("Unsafe mode enabled!")
	return container
}

func (container *di) parseBeanType(beanType interface{}) (prototype reflect.Type, beanName string) {
	prototype = reflect.Indirect(reflect.ValueOf(beanType)).Type()
	// 生成beanName
	tmpBeanName := reflect.New(prototype).Interface()
	switch tmpBeanName.(type) {
	case BeanName:
		if name := tmpBeanName.(BeanName).BeanName(); name != "" {
			container.log.Debug(fmt.Sprintf("beanName generate by interface BeanName for type %T, beanName: %s", beanType, name))
			beanName = name
		}
	}
	if beanName == "" {
		beanName = GetBeanName(beanType)
		container.log.Debug(fmt.Sprintf("beanName generate by default for type %T, beanName: %s", beanType, beanName))
	}
	return
}

func (container *di) DebugMode(enable bool) DI {
	container.log.DebugMode(enable)
	return container
}

func (container *di) Log(log Log) DI {
	container.log = log
	return container
}

// RegisterBean 注册一个已生成的bean，根据bean类型生成beanName
func (container *di) RegisterBean(bean interface{}) DI {
	return container.RegisterNamedBean("", bean)
}

// RegisterNamedBean 以指定名称注册一个bean
func (container *di) RegisterNamedBean(beanName string, bean interface{}) DI {
	if !IsPtr(bean) {
		container.log.Fatal(fmt.Sprintf("%s: bean must be a pointer", ErrBean))
		return container
	}
	if beanName == "" {
		_, beanName = container.parseBeanType(bean)
	}
	if _, exist := container.beanMap[beanName]; exist {
		container.log.Fatal(fmt.Sprintf("%s: bean %s already exists", ErrBean, beanName))
		return container
	}
	container.beanMap[beanName] = bean
	// 加入队列
	container.beanSort.PushBack(beanName)
	container.log.Info(fmt.Sprintf("register bean with name: %s", beanName))
	return container
}

func (container *di) Provide(prototype interface{}) DI {
	container.ProvideNamedBean("", prototype)
	return container
}

func (container *di) ProvideNamedBean(beanName string, beanType interface{}) DI {
	if container.loaded {
		container.log.Fatal(ErrLoaded.Error())
		return container
	}
	var prototype reflect.Type
	if beanName == "" {
		prototype, beanName = container.parseBeanType(beanType)
	} else {
		prototype, _ = container.parseBeanType(beanType)
	}

	// 检查bean重复
	if _, exist := container.beanMap[beanName]; exist {
		container.log.Fatal(fmt.Sprintf("%s: bean %s already exists", ErrBean, beanName))
		return container
	}
	// 检查beanDefinition重复
	if existDefinition, exist := container.beanDefinitionMap[beanName]; exist {
		container.log.Fatal(fmt.Sprintf("%s: bean %s already defined by %s", ErrDefinition, beanName, existDefinition.Type.String()))
		return container
	} else {
		container.beanDefinitionMap[beanName] = container.newDefinition(beanName, prototype)
		// 加入队列
		container.beanSort.PushBack(beanName)
	}
	container.log.Info(fmt.Sprintf("provide bean with name: %s", beanName))
	return container
}

func (container *di) GetBean(beanName string) (interface{}, bool) {
	bean, ok := container.beanMap[beanName]
	return bean, ok
}

func (container *di) GetByType(beanType interface{}) (interface{}, bool) {
	var typeValue reflect.Type
	if IsPtr(beanType) {
		typeValue = reflect.TypeOf(beanType)
	} else {
		typeValue = reflect.PtrTo(reflect.TypeOf(beanType))
	}
	for _, bean := range container.beanMap {
		if reflect.TypeOf(bean).AssignableTo(typeValue) {
			return bean, true
		}
	}
	return nil, false
}

func (container *di) NewBean(beanType interface{}) (bean interface{}) {
	prototype, beanName := container.parseBeanType(beanType)
	// 检查beanDefinition是否存在
	if _, exist := container.beanDefinitionMap[beanName]; !exist {
		return container.newBean(container.newDefinition(beanName, prototype))
	} else {
		return container.NewBeanByName(beanName)
	}
}

func (container *di) NewBeanByName(beanName string) (bean interface{}) {
	def, ok := container.beanDefinitionMap[beanName]
	if !ok {
		panic(fmt.Errorf("%w: %s notfound", ErrDefinition, beanName))
	}
	return container.newBean(def)
}

func (container *di) newBean(def definition) (bean interface{}) {
	container.log.Info(fmt.Sprintf("new bean instance %s", def.Name))
	// 反射实例并注入值
	prototype := container.instanceBean(def)
	// 触发构造方法
	container.constructBean(def.Name, prototype)
	// 触发注入 bean
	bean = container.processBean(prototype, def)
	// 初始化完成
	container.initializedBean(def.Name, bean)
	// 使用析构函数来完成 bean 的 destroy
	runtime.SetFinalizer(bean, func(bean interface{}) {
		container.destroyBean(def.Name, bean)
	})
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
	var cancel context.CancelFunc
	container.ctx, cancel = context.WithCancel(ctx)
	<-ctx.Done()
	defer cancel()
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
			container.constructBean(beanName, prototype)
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
			container.log.Info(fmt.Sprintf("initialize bean %s(%T)", def.Name, prototype))
			// 加载完成的bean放入beanMap中
			container.beanMap[beanName] = container.processBean(prototype, def)
		}
	}
}

// initialized 容器初始化完成
func (container *di) initialized() {
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		bean := container.beanMap[beanName]
		container.initializedBean(beanName, bean)
	}
}

func (container *di) destroyBeans() {
	// 倒序销毁bean
	for e := container.beanSort.Back(); e != nil; e = e.Prev() {
		beanName := e.Value.(string)
		if bean, ok := container.beanMap[beanName]; ok {
			container.destroyBean(beanName, bean)
			delete(container.beanMap, beanName)
		}
	}
}
