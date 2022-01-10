package di

import (
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"unsafe"
)

// wireValue 注入配置项
func (container *di) wireValue(bean reflect.Value, def definition, prefix string) {
	if len(def.valueMap) > 0 {
		container.log.Info(fmt.Sprintf("wire value for bean %s(%s)", def.Name, def.Type.String()))
	}
	for filedName, valueInfo := range def.valueMap {
		valueName := prefix + valueInfo.Name
		value := container.valueStore.Get(valueName)
		if value == nil {
			continue
		}
		castValue, err := van.Cast(value, valueInfo.Type)
		if err != nil {
			container.log.Fatal(fmt.Sprintf("%s: %s(%s) wire value failed for %s(%s.%s), %s",
				ErrBean, valueName, valueInfo.Type.String(),
				def.Name, def.Type.String(), filedName,
				err.Error(),
			))
			return
		}
		val := reflect.ValueOf(castValue)
		// 设置值
		if container.unsafe {
			container.log.Debug(fmt.Sprintf("wire value for %s(%s.%s) in unsafe mode",
				def.Name, def.Type.String(), filedName,
			))
			field := bean.FieldByName(filedName)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(val)
		} else {
			container.log.Debug(fmt.Sprintf("wire value for %s(%s.%s)",
				def.Name, def.Type.String(), filedName,
			))
			bean.FieldByName(filedName).Set(val)
		}
	}
}

// instanceBean 创建bean指针对象 并注入value
func (container *di) instanceBean(def definition) interface{} {
	container.log.Debug(fmt.Sprintf("reflect instance for %s(%s)", def.Name, def.Type.String()))
	prototype := reflect.New(def.Type).Interface()
	// 注入值
	container.wireValue(reflect.ValueOf(prototype).Elem(), def, "")
	return prototype
}

// constructBean 触发bean构造方法
func (container *di) constructBean(beanName string, prototype interface{}) {
	switch prototype.(type) {
	case BeanConstructWithContainer:
		container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstructWithContainer for %s(%T)", beanName, prototype))
		prototype.(BeanConstructWithContainer).BeanConstruct(container)
	case BeanConstruct:
		container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstruct for %s(%T)", beanName, prototype))
		prototype.(BeanConstruct).BeanConstruct()
	}
}

// processBean 处理bean依赖注入
func (container *di) processBean(prototype interface{}, def definition) interface{} {
	// 注入前方法
	switch prototype.(type) {
	case PreInitializeWithContainer:
		container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitializeWithContainer for %s(%s)", def.Name, def.Type.String()))
		prototype.(PreInitializeWithContainer).PreInitialize(container)
	case PreInitialize:
		container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitialize for %s(%s)", def.Name, def.Type.String()))
		prototype.(PreInitialize).PreInitialize()
	}

	bean := reflect.ValueOf(prototype).Elem()
	container.wireBean(bean, def)

	// 注入后方法
	switch prototype.(type) {
	case AfterPropertiesSetWithContainer:
		container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSetWithContainer for %s(%s)", def.Name, def.Type.String()))
		prototype.(AfterPropertiesSetWithContainer).AfterPropertiesSet(container)
	case AfterPropertiesSet:
		container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSet for %s(%s)", def.Name, def.Type.String()))
		prototype.(AfterPropertiesSet).AfterPropertiesSet()
	}
	return prototype
}

// findBeanByName 根据名称查找bean
func (container *di) findBeanByName(beanName string) (awareBean interface{}, ok bool) {
	// 从注册的bean中查找
	if awareBean, ok = container.beanMap[beanName]; !ok {
		// 从原型定义中查找
		awareBean, ok = container.prototypeMap[beanName]
	}
	return
}

func (container *di) findBeanByType(beanType reflect.Type) (bean interface{}, beanName string) {
	// 根据排序遍历beanName查找
	for e := container.beanSort.Front(); e != nil; e = e.Next() {
		findBeanName := e.Value.(string)
		if prototype, ok := container.prototypeMap[findBeanName]; ok {
			if reflect.TypeOf(prototype).AssignableTo(beanType) {
				container.log.Info(fmt.Sprintf("find interface %s implemented by %s(%T)",
					beanType.String(), findBeanName, prototype,
				))
				bean = prototype
				beanName = findBeanName
			}
		}
	}
	// TODO 起始可能会找到多个，但是这里按bean注册的先后顺序去处理，取最后一个
	return
}

// wireBean 注入单个依赖
func (container *di) wireBean(bean reflect.Value, def definition) {
	if len(def.awareMap) > 0 {
		container.log.Info(fmt.Sprintf("wire field for bean %s(%s)", def.Name, def.Type.String()))
	}
	for filedName, awareInfo := range def.awareMap {
		var awareBean interface{}
		var ok bool

		// 根据名称查找bean
		awareBean, ok = container.findBeanByName(awareInfo.Name)
		// 如果是接口类型
		if awareInfo.IsInterface && !ok {
			var interfaceBeanName string
			awareBean, interfaceBeanName = container.findBeanByType(awareInfo.Type)
			if awareBean != nil {
				ok = true
				container.log.Info(fmt.Sprintf("%s(%T) will be set to %s(%s.%s)",
					interfaceBeanName, awareBean,
					def.Name, def.Type.String(), filedName,
				))
			}
		}
		if !ok {
			if awareInfo.Omitempty {
				container.log.Warn(fmt.Sprintf("Omitempty: dependent bean %s not found for %s(%s.%s)",
					awareInfo.Name,
					def.Name,
					def.Type.String(),
					filedName))
				continue
			}
			container.log.Fatal(fmt.Sprintf("%s: %s notfound for %s(%s.%s)",
				ErrBean,
				awareInfo.Name,
				def.Name,
				def.Type.String(),
				filedName))
		}
		value := reflect.ValueOf(awareBean)

		// 类型检查
		if awareInfo.IsPtr { // 指针类型
			if !value.Type().AssignableTo(awareInfo.Type) {
				container.log.Fatal(fmt.Sprintf("%s: %s(%s) not match for %s(%s.%s) need type %s",
					ErrBean,
					awareInfo.Name, value.Type().String(),
					def.Name,
					def.Type.String(),
					filedName,
					awareInfo.Type.String(),
				))
				return
			}
		} else { // 接口类型
			if !value.Type().Implements(awareInfo.Type) {
				container.log.Fatal(fmt.Sprintf("%s: %s(%s) not implements interface %s for %s(%s.%s)",
					ErrBean,
					awareInfo.Name, value.Type().String(),
					awareInfo.Type.String(),
					def.Name,
					def.Type.String(),
					filedName,
				))
				return
			}
		}

		// 设置值
		if container.unsafe {
			if awareInfo.Anonymous {
				container.log.Debug(fmt.Sprintf("wire anonymous field for %s(%s.%s) in unsafe mode",
					def.Name, def.Type.String(), filedName,
				))
			} else {
				container.log.Debug(fmt.Sprintf("wire field for %s(%s.%s) in unsafe mode",
					def.Name, def.Type.String(), filedName,
				))
			}

			field := bean.FieldByName(filedName)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(value)
		} else {
			if awareInfo.Anonymous {
				container.log.Debug(fmt.Sprintf("wire anonymous field for %s(%s.%s)",
					def.Name, def.Type.String(), filedName,
				))
			} else {
				container.log.Debug(fmt.Sprintf("wire field for %s(%s.%s)",
					def.Name, def.Type.String(), filedName,
				))
			}

			bean.FieldByName(filedName).Set(value)
		}
	}
}

// processInitialized bean初始化完成
func (container *di) initializedBean(beanName string, bean interface{}) {
	switch bean.(type) {
	case InitializedWithContainer:
		container.log.Debug(fmt.Sprintf("call lifecycle interface InitializedWithContainer for %s(%T)", beanName, bean))
		bean.(InitializedWithContainer).Initialized(container)
	case Initialized:
		container.log.Debug(fmt.Sprintf("call lifecycle interface Initialized for %s(%T)", beanName, bean))
		bean.(Initialized).Initialized()
	}
}

// destroyBean 销毁bean
func (container *di) destroyBean(beanName string, bean interface{}) {
	switch bean.(type) {
	case DisposableWithContainer:
		container.log.Debug(fmt.Sprintf("call lifecycle interface DisposableWithContainer for %s(%T)", beanName, bean))
		bean.(DisposableWithContainer).Destroy(container)
	case Disposable:
		container.log.Debug(fmt.Sprintf("call lifecycle interface Disposable for %s(%T)", beanName, bean))
		bean.(Disposable).Destroy()
	}
}
