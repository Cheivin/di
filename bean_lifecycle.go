package di

import (
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"unsafe"
)

// wireValue 注入配置项
func (container *di) wireValue(bean reflect.Value, def definition) {
	if len(def.valueMap) > 0 {
		container.log.Info(fmt.Sprintf("wire value for bean %s(%s)", def.Name, def.Type.String()))
	}
	for filedName, valueInfo := range def.valueMap {
		value := container.valueStore.Get(valueInfo.Name)
		if value != nil {
			castValue, err := van.Cast(value, valueInfo.Type)
			if err != nil {
				container.log.Fatal(fmt.Sprintf("%s: %s(%s) wire value failed for %s(%s.%s), %s",
					ErrBean, valueInfo.Name, valueInfo.Type.String(),
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
}

// instanceBean 创建bean指针对象 并注入value
func (container *di) instanceBean(def definition) interface{} {
	container.log.Debug(fmt.Sprintf("reflect instance for %s(%s)", def.Name, def.Type.String()))
	prototype := reflect.New(def.Type).Interface()
	// 注入值
	container.wireValue(reflect.ValueOf(prototype).Elem(), def)
	return prototype
}

// constructBean 触发bean构造方法
func (container *di) constructBean(beanName string, prototype interface{}) {
	if construct, ok := prototype.(BeanConstructWithContainer); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstructWithContainer for %s(%T)", beanName, prototype))
		construct.BeanConstruct(container)
	} else if construct, ok := prototype.(BeanConstruct); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstruct for %s(%T)", beanName, prototype))
		construct.BeanConstruct()
	}
}

// processBean 处理bean依赖注入
func (container *di) processBean(prototype interface{}, def definition) interface{} {
	// 注入前方法
	if initialize, ok := prototype.(PreInitializeWithContainer); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitializeWithContainer for %s(%s)", def.Name, def.Type.String()))
		initialize.PreInitialize(container)
	} else if initialize, ok := prototype.(PreInitialize); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitialize for %s(%s)", def.Name, def.Type.String()))
		initialize.PreInitialize()
	}
	bean := reflect.ValueOf(prototype).Elem()
	container.wireBean(bean, def)
	// 注入后方法
	if propertiesSet, ok := prototype.(AfterPropertiesSetWithContainer); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSetWithContainer for %s(%s)", def.Name, def.Type.String()))
		propertiesSet.AfterPropertiesSet(container)
	} else if propertiesSet, ok := prototype.(AfterPropertiesSet); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSet for %s(%s)", def.Name, def.Type.String()))
		propertiesSet.AfterPropertiesSet()
	}
	return prototype
}

// wireBean 注入单个依赖
func (container *di) wireBean(bean reflect.Value, def definition) {
	if len(def.awareMap) > 0 {
		container.log.Info(fmt.Sprintf("wire field for bean %s(%s)", def.Name, def.Type.String()))
	}
	for filedName, awareInfo := range def.awareMap {
		var awareBean interface{}
		var ok bool
		if awareBean, ok = container.beanMap[awareInfo.Name]; !ok {
			// 手动注册的bean中找不到，尝试查找原型定义
			if awareBean, ok = container.prototypeMap[awareInfo.Name]; !ok {
				// 如果原型定义找不到，判断是否Omitempty
				if awareInfo.Omitempty {
					container.log.Warn(fmt.Sprintf("Omitempty: dependent bean %s not found for %s(%s.%s)",
						awareInfo.Name,
						def.Name,
						def.Type.String(),
						filedName,
					))
					continue
				}
				container.log.Fatal(fmt.Sprintf("%s: %s notfound for %s(%s.%s)",
					ErrBean,
					awareInfo.Name,
					def.Name,
					def.Type.String(),
					filedName,
				))
				return
			}
		}
		value := reflect.ValueOf(awareBean)
		// 匿名字段不能实现BeanConstruct/PreInitialize/AfterPropertiesSet/Initialized/Disposable等生命周期接口
		if awareInfo.Anonymous {
			errMsg := "%s: %s(%s) as anonymous field in %s(%s.%s) can not implements %s"
			if _, ok := awareBean.(BeanConstruct); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"BeanConstruct",
				))
				return
			} else if _, ok := awareBean.(BeanConstructWithContainer); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"BeanConstructWithContainer",
				))
				return
			}
			if _, ok := awareBean.(PreInitialize); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"PreInitialize",
				))
				return
			} else if _, ok := awareBean.(PreInitializeWithContainer); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"PreInitializeWithContainer",
				))
				return
			}
			if _, ok := awareBean.(AfterPropertiesSet); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"AfterPropertiesSet",
				))
				return
			} else if _, ok := awareBean.(AfterPropertiesSetWithContainer); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"AfterPropertiesSetWithContainer",
				))
				return
			}
			if _, ok := awareBean.(Initialized); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"Initialized",
				))
				return
			} else if _, ok := awareBean.(InitializedWithContainer); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"InitializedWithContainer",
				))
				return
			}
			if _, ok := awareBean.(Disposable); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"Disposable",
				))
				return
			} else if _, ok := awareBean.(DisposableWithContainer); ok {
				container.log.Fatal(fmt.Sprintf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"DisposableWithContainer",
				))
				return
			}
		}
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
	// 初始化完成
	if initialized, ok := bean.(InitializedWithContainer); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface InitializedWithContainer for %s(%T)", beanName, bean))

		initialized.Initialized(container)
	} else if initialized, ok := bean.(Initialized); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface Initialized for %s(%T)", beanName, bean))

		initialized.Initialized()
	}
}

// destroyBean 销毁bean
func (container *di) destroyBean(beanName string, bean interface{}) {
	if disposable, ok := bean.(DisposableWithContainer); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface DisposableWithContainer for %s(%T)", beanName, bean))

		disposable.Destroy(container)
	} else if disposable, ok := bean.(Disposable); ok {
		container.log.Debug(fmt.Sprintf("call lifecycle interface Disposable for %s(%T)", beanName, bean))

		disposable.Destroy()
	}
}
