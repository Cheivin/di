package di

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"unsafe"
)

type (
	DI struct {
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
	ErrLoaded     = errors.New("DI loaded")
)

func New() *DI {
	return &DI{
		beanDefinitionMap: map[string]definition{},
		prototypeMap:      map[string]interface{}{},
		beanMap:           map[string]interface{}{},
		valueStore:        van.New(),
		beanSort:          list.New(),
	}
}

func (di *DI) UnsafeMode(open bool) *DI {
	di.unsafe = open
	return di
}

// RegisterBean 注册一个已生成的bean，根据bean类型生成beanName
func (di *DI) RegisterBean(bean interface{}) *DI {
	return di.RegisterNamedBean("", bean)
}

// RegisterNamedBean 以指定名称注册一个bean
func (di *DI) RegisterNamedBean(beanName string, bean interface{}) *DI {
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
	if _, exist := di.beanMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
	}
	di.beanMap[beanName] = bean
	// 加入队列
	di.beanSort.PushBack(beanName)
	return di
}

func (di *DI) Provide(prototype interface{}) *DI {
	di.ProvideNamedBean("", prototype)
	return di
}

func (di *DI) ProvideNamedBean(beanName string, beanType interface{}) *DI {
	if di.loaded {
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
	if _, exist := di.beanMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
	}
	// 检查beanDefinition重复
	if existDefinition, exist := di.beanDefinitionMap[beanName]; exist {
		panic(fmt.Errorf("%w: bean %s already defined by %s", ErrDefinition, beanName, existDefinition.Type.String()))
	} else {
		di.beanDefinitionMap[beanName] = newDefinition(beanName, prototype)
		// 加入队列
		di.beanSort.PushBack(beanName)
	}
	return di
}

func (di *DI) GetBean(beanName string) (interface{}, bool) {
	bean, ok := di.beanMap[beanName]
	return bean, ok
}

func (di *DI) Load() {
	if di.loaded {
		panic(ErrLoaded)
	}

	di.loaded = true
	di.initializeBeans()
	di.processBeans()
	di.initialized()

}

// initializeBeans 初始化bean对象
func (di *DI) initializeBeans() {
	// 创建类型的指针对象
	for beanName, def := range di.beanDefinitionMap {
		prototype := reflect.New(def.Type).Interface()
		di.prototypeMap[beanName] = prototype
		// 注入值
		di.wireValue(beanName, reflect.ValueOf(prototype).Elem(), def)
	}
	// 根据排序遍历触发BeanConstruct方法
	for e := di.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		if prototype, ok := di.prototypeMap[beanName]; ok {
			if construct, ok := prototype.(BeanConstructWithContainer); ok {
				construct.BeanConstruct(di)
			} else if construct, ok := prototype.(BeanConstruct); ok {
				construct.BeanConstruct()
			}
		}
	}
}

// processBeans 注入依赖
func (di *DI) processBeans() {
	for e := di.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		if prototype, ok := di.prototypeMap[beanName]; ok {
			def := di.beanDefinitionMap[beanName]
			di.processBean(beanName, prototype, def)
		}
	}
}

// processBean 处理bean
func (di *DI) processBean(beanName string, prototype interface{}, def definition) {
	// 注入前方法
	if initialize, ok := prototype.(PreInitializeWithContainer); ok {
		initialize.PreInitialize(di)
	} else if initialize, ok := prototype.(PreInitialize); ok {
		initialize.PreInitialize()
	}
	bean := reflect.ValueOf(prototype).Elem()
	di.wireBean(beanName, bean, def)
	// 注入后方法
	if propertiesSet, ok := prototype.(AfterPropertiesSetWithContainer); ok {
		propertiesSet.AfterPropertiesSet(di)
	} else if propertiesSet, ok := prototype.(AfterPropertiesSet); ok {
		propertiesSet.AfterPropertiesSet()
	}
	// 加载为bean
	di.beanMap[beanName] = prototype
}

// processBeans 注入依赖
func (di *DI) wireBean(beanName string, bean reflect.Value, def definition) {
	for filedName, awareInfo := range def.awareMap {
		var awareBean interface{}
		var ok bool
		if awareBean, ok = di.beanMap[awareInfo.Name]; !ok {
			// 手动注册的bean中找不到，尝试查找原型定义
			if awareBean, ok = di.prototypeMap[awareInfo.Name]; !ok {
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
		if di.unsafe {
			field := bean.FieldByName(filedName)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(value)
		} else {
			bean.FieldByName(filedName).Set(value)
		}
	}
}

// wireValue 注入配置项
func (di *DI) wireValue(beanName string, bean reflect.Value, def definition) {
	for filedName, valueInfo := range def.valueMap {
		value := di.valueStore.Get(valueInfo.Name)
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
			if di.unsafe {
				field := bean.FieldByName(filedName)
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
				field.Set(val)
			} else {
				bean.FieldByName(filedName).Set(val)
			}
		}
	}
}

func (di *DI) initialized() {
	for e := di.beanSort.Front(); e != nil; e = e.Next() {
		beanName := e.Value.(string)
		bean := di.beanMap[beanName]
		// 初始化完成
		if initialized, ok := bean.(InitializedWithContainer); ok {
			initialized.Initialized(di)
		} else if initialized, ok := bean.(Initialized); ok {
			initialized.Initialized()
		}
	}
}

func (di *DI) destroyBean(beanName string) {
	if bean, ok := di.beanMap[beanName]; ok {
		if disposable, ok := bean.(DisposableWithContainer); ok {
			disposable.Destroy(di)
		} else if disposable, ok := bean.(Disposable); ok {
			disposable.Destroy()
		}
		delete(di.beanMap, beanName)
	}
}

func (di *DI) destroyBeans() {
	// 倒序销毁bean
	for e := di.beanSort.Back(); e != nil; e = e.Prev() {
		di.destroyBean(e.Value.(string))
	}
}

func (di *DI) Serve(ctx context.Context) {
	if !di.loaded {
		panic(ErrLoaded)
	}
	<-ctx.Done()
	di.destroyBeans()
}

func (di *DI) LoadAndServ(ctx context.Context) {
	di.Load()
	di.Serve(ctx)
}
