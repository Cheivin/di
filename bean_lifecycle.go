package di

import (
	"fmt"
	"github.com/cheivin/di/van"
	"reflect"
	"unsafe"
)

// wireValue 注入配置项
func (container *di) wireValue(bean reflect.Value, def definition) {
	for filedName, valueInfo := range def.valueMap {
		value := container.valueStore.Get(valueInfo.Name)
		if value != nil {
			castValue, err := van.Cast(value, valueInfo.Type)
			if err != nil {
				panic(fmt.Errorf("%w: %s(%s) wire value failed for %s(%s.%s), %s",
					ErrBean,
					valueInfo.Name, valueInfo.Type.String(),
					def.Name,
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

// instanceBean 创建bean指针对象 并注入value
func (container *di) instanceBean(def definition) interface{} {
	prototype := reflect.New(def.Type).Interface()
	// 注入值
	container.wireValue(reflect.ValueOf(prototype).Elem(), def)
	return prototype
}

// constructBean 触发bean构造方法
func (container *di) constructBean(prototype interface{}) {
	if construct, ok := prototype.(BeanConstructWithContainer); ok {
		construct.BeanConstruct(container)
	} else if construct, ok := prototype.(BeanConstruct); ok {
		construct.BeanConstruct()
	}
}

// processBean 处理bean依赖注入
func (container *di) processBean(prototype interface{}, def definition) interface{} {
	// 注入前方法
	if initialize, ok := prototype.(PreInitializeWithContainer); ok {
		initialize.PreInitialize(container)
	} else if initialize, ok := prototype.(PreInitialize); ok {
		initialize.PreInitialize()
	}
	bean := reflect.ValueOf(prototype).Elem()
	container.wireBean(bean, def)
	// 注入后方法
	if propertiesSet, ok := prototype.(AfterPropertiesSetWithContainer); ok {
		propertiesSet.AfterPropertiesSet(container)
	} else if propertiesSet, ok := prototype.(AfterPropertiesSet); ok {
		propertiesSet.AfterPropertiesSet()
	}
	return prototype
}

// wireBean 注入单个依赖
func (container *di) wireBean(bean reflect.Value, def definition) {
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
					def.Name,
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
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"BeanConstruct",
				))
			} else if _, ok := awareBean.(BeanConstructWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"BeanConstructWithContainer",
				))
			}
			if _, ok := awareBean.(PreInitialize); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"PreInitialize",
				))
			} else if _, ok := awareBean.(PreInitializeWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"PreInitializeWithContainer",
				))
			}
			if _, ok := awareBean.(AfterPropertiesSet); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"AfterPropertiesSet",
				))
			} else if _, ok := awareBean.(AfterPropertiesSetWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"AfterPropertiesSetWithContainer",
				))
			}
			if _, ok := awareBean.(Initialized); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"Initialized",
				))
			} else if _, ok := awareBean.(InitializedWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"InitializedWithContainer",
				))
			}
			if _, ok := awareBean.(Disposable); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
					"Disposable",
				))
			} else if _, ok := awareBean.(DisposableWithContainer); ok {
				panic(fmt.Errorf(errMsg,
					ErrBean, awareInfo.Name,
					value.Type().String(), def.Name, def.Type.String(), filedName,
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
					def.Name,
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
					def.Name,
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

// processInitialized bean初始化完成
func (container *di) initializedBean(bean interface{}) {
	// 初始化完成
	if initialized, ok := bean.(InitializedWithContainer); ok {
		initialized.Initialized(container)
	} else if initialized, ok := bean.(Initialized); ok {
		initialized.Initialized()
	}
}

// destroyBean 销毁bean
func (container *di) destroyBean(bean interface{}) {
	if disposable, ok := bean.(DisposableWithContainer); ok {
		disposable.Destroy(container)
	} else if disposable, ok := bean.(Disposable); ok {
		disposable.Destroy()
	}
}
