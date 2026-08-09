package di

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/cheivin/di/van"
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
			container.log.Fatal(fmt.Errorf("%w: %s(%s) wire value failed for %s(%s.%s), %s",
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
func (container *di) instanceBean(def definition) any {
	// 工厂模式：按入参类型注入依赖并调用工厂
	if def.factory.IsValid() {
		args := make([]reflect.Value, len(def.factoryArgs))
		for i, argType := range def.factoryArgs {
			argBean := container.resolveFactoryArg(argType)
			if argBean == nil {
				container.log.Fatal(fmt.Errorf("%w: factory arg %s notfound for %s",
					ErrBean, argType.String(), def.Name))
				return nil
			}
			args[i] = reflect.ValueOf(argBean)
		}
		results := def.factory.Call(args)
		container.log.Debug(fmt.Sprintf("factory instance for %s(%s)", def.Name, def.Type.String()))
		return results[0].Interface()
	}
	container.log.Debug(fmt.Sprintf("reflect instance for %s(%s)", def.Name, def.Type.String()))
	prototype := reflect.New(def.Type).Interface()
	// 注入值
	container.wireValue(reflect.ValueOf(prototype).Elem(), def, "")
	return prototype
}

// resolveFactoryArg 按类型解析工厂入参：先按类型名查找，找不到则按类型匹配候选
func (container *di) resolveFactoryArg(argType reflect.Type) any {
	// 指针类型：按类型推断 beanName 查找
	if argType.Kind() == reflect.Pointer {
		beanName := GetBeanName(argType)
		if bean, ok := container.findBeanByName(beanName); ok {
			return bean
		}
	}
	// 接口或按名未命中：按类型匹配（取 selector 选中的）
	candidates := container.findBeanByType(argType)
	if len(candidates) > 0 {
		idx, err := container.selector.Select(candidates, argType)
		if err != nil {
			container.log.Fatal(fmt.Errorf("%w: factory arg select failed for %s, %s",
				ErrBean, argType.String(), err.Error()))
			return nil
		}
		return candidates[idx].Bean
	}
	return nil
}

// processBean 处理单个 bean 的依赖注入流程：
// PreInitialize → wireBean（注入 aware 依赖）→ AfterPropertiesSet。
// 注入和回调都在锁外执行，允许 bean 回调内反向访问容器。
func (container *di) processBean(prototype any, def definition) any {
	// 注入前方法
	container.preInitialize(def, prototype)

	bean := reflect.ValueOf(prototype).Elem()
	container.wireBean(bean, def)

	// 注入后方法
	container.afterPropertiesSet(def, prototype)
	return prototype
}

// findBeanByName 根据名称查找bean
func (container *di) findBeanByName(beanName string) (awareBean any, ok bool) {
	container.mu.RLock()
	defer container.mu.RUnlock()
	// 从注册的bean中查找
	if awareBean, ok = container.beanMap[beanName]; !ok {
		// 从原型定义中查找
		awareBean, ok = container.prototypeMap[beanName]
	}
	return
}

type BeanWithName struct {
	Name string
	Bean any
}

// findBeanByType 按 beanSort 注册顺序收集所有可赋值给 beanType 的 bean。
// 锁内读取 beanMap/prototypeMap 快照，锁外打印日志，避免持锁调日志的潜在重入。
// 不调用 findBeanByName 以免重复加锁。
func (container *di) findBeanByType(beanType reflect.Type) []BeanWithName {
	var beans []BeanWithName
	// 根据排序遍历beanName查找
	container.mu.RLock()
	for _, findBeanName := range container.beanSort {
		// 锁内读取 prototype/bean（不调用 findBeanByName 以免重复加锁）
		prototype, ok := container.beanMap[findBeanName]
		if !ok {
			prototype, ok = container.prototypeMap[findBeanName]
		}
		if ok {
			if reflect.TypeOf(prototype).AssignableTo(beanType) {
				beans = append(beans, BeanWithName{Name: findBeanName, Bean: prototype})
			}
		}
	}
	container.mu.RUnlock()
	// 日志在锁外
	for _, b := range beans {
		container.log.Info(fmt.Sprintf("find interface %s implemented by %s(%T)",
			beanType.String(), b.Name, b.Bean,
		))
	}
	return beans
}

// wireBean 注入单个依赖
func (container *di) wireBean(bean reflect.Value, def definition) {
	if len(def.awareMap) > 0 {
		container.log.Info(fmt.Sprintf("wire field for bean %s(%s)", def.Name, def.Type.String()))
	}
	for filedName, awareInfo := range def.awareMap {
		// slice/map 批量注入：收集所有可赋值给元素类型的 bean，不走单值选择
		if awareInfo.IsSlice || awareInfo.IsMap {
			candidates := container.findBeanByType(awareInfo.ElemType)
			field := bean.FieldByName(filedName)
			if container.unsafe {
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}
			if awareInfo.IsSlice {
				sliceVal := reflect.MakeSlice(awareInfo.Type, len(candidates), len(candidates))
				for i, c := range candidates {
					sliceVal.Index(i).Set(reflect.ValueOf(c.Bean))
				}
				field.Set(sliceVal)
				container.log.Debug(fmt.Sprintf("wire slice field for %s(%s.%s) with %d beans",
					def.Name, def.Type.String(), filedName, len(candidates)))
			} else { // map[string]T
				mapVal := reflect.MakeMapWithSize(awareInfo.Type, len(candidates))
				for _, c := range candidates {
					mapVal.SetMapIndex(reflect.ValueOf(c.Name), reflect.ValueOf(c.Bean))
				}
				field.Set(mapVal)
				container.log.Debug(fmt.Sprintf("wire map field for %s(%s.%s) with %d beans",
					def.Name, def.Type.String(), filedName, len(candidates)))
			}
			continue
		}

		var awareBean any
		var ok bool

		// 根据名称查找bean
		awareBean, ok = container.findBeanByName(awareInfo.Name)
		// 如果是接口类型
		if awareInfo.IsInterface && !ok {
			awareBeans := container.findBeanByType(awareInfo.Type)
			if len(awareBeans) > 0 {
				idx, err := container.selector.Select(awareBeans, awareInfo.Type)
				if err != nil {
					container.log.Fatal(fmt.Errorf("%w: select failed for %s(%s.%s), %s",
						ErrBean, def.Name, def.Type.String(), filedName, err.Error()))
					return
				}
				selectBean := awareBeans[idx]
				awareBean = selectBean.Bean
				ok = true
				container.log.Info(fmt.Sprintf("%s(%T) will be set to %s(%s.%s)",
					selectBean.Name, awareBean,
					def.Name, def.Type.String(), filedName,
				))
			}
		}

		injectInfo := &InjectInfo{
			Bean:        awareBean,
			BeanName:    awareInfo.Name,
			Type:        awareInfo.Type,
			IsPtr:       awareInfo.IsPtr,
			IsInterface: awareInfo.IsInterface,
			Anonymous:   awareInfo.Anonymous,
			Omitempty:   awareInfo.Omitempty,
			IsSlice:     awareInfo.IsSlice,
			IsMap:       awareInfo.IsMap,
			ElemType:    awareInfo.ElemType,
		}
		switch bean.Interface().(type) {
		case Injector:
			bean.Interface().(Injector).BeanInject(container, injectInfo)
			if !ok {
				ok = injectInfo.Bean != nil
			}
			awareBean = injectInfo.Bean
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
			container.log.Fatal(fmt.Errorf("%w: %s notfound for %s(%s.%s)",
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
				container.log.Fatal(fmt.Errorf("%w: %s(%s) not match for %s(%s.%s) need type %s",
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
				container.log.Fatal(fmt.Errorf("%w: %s(%s) not implements interface %s for %s(%s.%s)",
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
