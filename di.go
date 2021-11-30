package di

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"runtime"
	"unsafe"
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

func (container *di) ProvideNamedBean(beanName string, beanType interface{}) DI {
	if container.loaded {
		panic(ErrLoaded)
	}
	var prototype reflect.Type
	if IsPtr(beanType) {
		prototype = reflect.TypeOf(beanType).Elem()
	} else {
		prototype = reflect.TypeOf(beanType)
	}
	if beanName == "" {
		if tmpBeanName, ok := (reflect.New(prototype).Interface()).(BeanName); ok {
			if name := tmpBeanName.BeanName(); name != "" {
				beanName = name
			} else {
				beanName = GetBeanName(beanType)
			}
		} else {
			beanName = GetBeanName(beanType)
		}
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

func (container *di) Load() {
	if container.loaded {
		panic(ErrLoaded)
	}

	container.loaded = true
	container.initializeBeans()
	container.processBeans()
	container.initialized()

}

func (container *di) NewBean(beanType interface{}) (bean interface{}) {
	var prototype reflect.Type
	if IsPtr(beanType) {
		prototype = reflect.TypeOf(beanType).Elem()
	} else {
		prototype = reflect.TypeOf(beanType)
	}
	var beanName string
	if tmpBeanName, ok := (reflect.New(prototype).Interface()).(BeanName); ok {
		if name := tmpBeanName.BeanName(); name != "" {
			beanName = name
		} else {
			beanName = GetBeanName(beanType)
		}
	} else {
		beanName = GetBeanName(beanType)
	}
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
	container.wireValue(beanName, reflect.ValueOf(prototype).Elem(), def)
	// 触发BeanConstruct
	if construct, ok := prototype.(BeanConstructWithContainer); ok {
		construct.BeanConstruct(container)
	} else if construct, ok := prototype.(BeanConstruct); ok {
		construct.BeanConstruct()
	}
	// 触发注入 bean
	bean = container.processBean(beanName, prototype, def)
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

// initializeBeans 初始化bean对象
func (container *di) initializeBeans() {
	// 创建类型的指针对象
	for beanName, def := range container.beanDefinitionMap {
		prototype := reflect.New(def.Type).Interface()
		container.prototypeMap[beanName] = prototype
		// 注入值
		container.wireValue(beanName, reflect.ValueOf(prototype).Elem(), def)
	}
	// 根据排序遍历触发BeanConstruct方法
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		if prototype, ok := container.prototypeMap[beanName]; ok {
			if construct, ok := prototype.(BeanConstructWithContainer); ok {
				construct.BeanConstruct(container)
			} else if construct, ok := prototype.(BeanConstruct); ok {
				construct.BeanConstruct()
			}
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
			container.beanMap[beanName] = container.processBean(beanName, prototype, def)
		}
	}
}

// processBean 处理bean
func (container *di) processBean(beanName string, prototype interface{}, def definition) interface{} {
	// 注入前方法
	if initialize, ok := prototype.(PreInitializeWithContainer); ok {
		initialize.PreInitialize(container)
	} else if initialize, ok := prototype.(PreInitialize); ok {
		initialize.PreInitialize()
	}
	bean := reflect.ValueOf(prototype).Elem()
	container.wireBean(beanName, bean, def)
	// 注入后方法
	if propertiesSet, ok := prototype.(AfterPropertiesSetWithContainer); ok {
		propertiesSet.AfterPropertiesSet(container)
	} else if propertiesSet, ok := prototype.(AfterPropertiesSet); ok {
		propertiesSet.AfterPropertiesSet()
	}
	return prototype
}

// processBeans 注入依赖
func (container *di) wireBean(beanName string, bean reflect.Value, def definition) {
	for filedName, awareInfo := range def.awareMap {
		var awareBean interface{}
		var ok bool
		if awareBean, ok = container.beanMap[awareInfo.Name]; !ok {
			// 手动注册的bean中找不到，尝试查找原型定义
			if awareBean, ok = container.prototypeMap[awareInfo.Name]; !ok {
				// 如果原型定义找不到，判断是否Omitempty
				if awareInfo.Omitempty {
					continue
				}
				panic(fmt.Errorf("%w: %s notfound for %s(%s.%s)",
					ErrBean,
					awareInfo.Name,
					beanName,
					def.Type.String(),
					filedName,
				))
			}
		}
		value := reflect.ValueOf(awareBean)
		// 匿名字段不能实现BeanConstruct/PreInitialize/AfterPropertiesSet/Initialized/Disposable等生命周期接口
		if awareInfo.Anonymous {
			errMsg := "%w: %s(%s) as anonymous field in %s(%s.%s) can not implements %s"
			if _, ok := awareBean.(BeanConstruct); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"BeanConstruct",
				))
			} else if _, ok := awareBean.(BeanConstructWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"BeanConstructWithContainer",
				))
			}
			if _, ok := awareBean.(PreInitialize); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"PreInitialize",
				))
			} else if _, ok := awareBean.(PreInitializeWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"PreInitializeWithContainer",
				))
			}
			if _, ok := awareBean.(AfterPropertiesSet); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"AfterPropertiesSet",
				))
			} else if _, ok := awareBean.(AfterPropertiesSetWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"AfterPropertiesSetWithContainer",
				))
			}
			if _, ok := awareBean.(Initialized); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"Initialized",
				))
			} else if _, ok := awareBean.(InitializedWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"InitializedWithContainer",
				))
			}
			if _, ok := awareBean.(Disposable); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"Disposable",
				))
			} else if _, ok := awareBean.(DisposableWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), beanName, def.Type.String(), filedName,
					"DisposableWithContainer",
				))
			}
		}
		// 类型检查
		if awareInfo.IsPtr { // 指针类型
			if !value.Type().AssignableTo(awareInfo.Type) {
				panic(fmt.Errorf("%w: %s(%s) not match for %s(%s.%s) need type %s",
					ErrBean,
					awareInfo.Name, value.Type().String(),
					beanName,
					def.Type.String(),
					filedName,
					awareInfo.Type.String(),
				))
			}
		} else { // 接口类型
			if !value.Type().Implements(awareInfo.Type) {
				panic(fmt.Errorf("%w: %s(%s) not implements interface %s for %s(%s.%s)",
					ErrBean,
					awareInfo.Name, value.Type().String(),
					awareInfo.Type.String(),
					beanName,
					def.Type.String(),
					filedName,
				))
			}
		}

		// 设置值
		if container.unsafe {
			field := bean.FieldByName(filedName)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(value)
		} else {
			bean.FieldByName(filedName).Set(value)
		}
	}
}

// wireValue 注入配置项
func (container *di) wireValue(beanName string, bean reflect.Value, def definition) {
	for filedName, valueInfo := range def.valueMap {
		value := container.valueStore.Get(valueInfo.Name)
		if value != nil {
			castValue, err := van.Cast(value, valueInfo.Type)
			if err != nil {
				panic(fmt.Errorf("%w: %s(%s) wire value failed for %s(%s.%s), %s",
					ErrBean,
					valueInfo.Name, valueInfo.Type.String(),
					beanName,
					def.Type.String(),
					filedName,
					err.Error(),
				))
			}
			val := reflect.ValueOf(castValue)
			// 设置值
			if container.unsafe {
				field := bean.FieldByName(filedName)
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
				field.Set(val)
			} else {
				bean.FieldByName(filedName).Set(val)
			}
		}
	}
}

func (container *di) initialized() {
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		bean := container.beanMap[beanName]
		// 初始化完成
		if initialized, ok := bean.(InitializedWithContainer); ok {
			initialized.Initialized(container)
		} else if initialized, ok := bean.(Initialized); ok {
			initialized.Initialized()
		}
	}
}

func (container *di) destroyBean(bean interface{}) {
	fmt.Println("清除 bean", bean)
	if disposable, ok := bean.(DisposableWithContainer); ok {
		disposable.Destroy(container)
	} else if disposable, ok := bean.(Disposable); ok {
		disposable.Destroy()
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

func (container *di) Serve(ctx context.Context) {
	if !container.loaded {
		panic(ErrLoaded)
	}
	<-ctx.Done()
	container.destroyBeans()
}
