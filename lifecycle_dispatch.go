package di

import (
	"fmt"
)

// dispatchLifecycle 按接口优先级分发生命周期回调：优先 WithContainer 版本，否则 plain 版本。
// 命中任一版本返回 true。两者都未实现返回 false。
func dispatchLifecycle[T any, W any](
	bean any,
	withContainer func(W),
	plain func(T),
) bool {
	if v, ok := bean.(W); ok {
		withContainer(v)
		return true
	}
	if v, ok := bean.(T); ok {
		plain(v)
		return true
	}
	return false
}

func (container *di) constructBean(beanName string, prototype any) {
	dispatchLifecycle(prototype,
		func(v BeanConstructWithContainer) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstructWithContainer for %s(%T)", beanName, prototype))
			v.BeanConstruct(container)
		},
		func(v BeanConstruct) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface BeanConstruct for %s(%T)", beanName, prototype))
			v.BeanConstruct()
		},
	)
}

func (container *di) preInitialize(def definition, prototype any) {
	dispatchLifecycle(prototype,
		func(v PreInitializeWithContainer) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitializeWithContainer for %s(%s)", def.Name, def.Type.String()))
			v.PreInitialize(container)
		},
		func(v PreInitialize) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface PreInitialize for %s(%s)", def.Name, def.Type.String()))
			v.PreInitialize()
		},
	)
}

func (container *di) afterPropertiesSet(def definition, prototype any) {
	dispatchLifecycle(prototype,
		func(v AfterPropertiesSetWithContainer) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSetWithContainer for %s(%s)", def.Name, def.Type.String()))
			v.AfterPropertiesSet(container)
		},
		func(v AfterPropertiesSet) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface AfterPropertiesSet for %s(%s)", def.Name, def.Type.String()))
			v.AfterPropertiesSet()
		},
	)
}

func (container *di) initializedBean(beanName string, bean any) {
	dispatchLifecycle(bean,
		func(v InitializedWithContainer) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface InitializedWithContainer for %s(%T)", beanName, bean))
			v.Initialized(container)
		},
		func(v Initialized) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface Initialized for %s(%T)", beanName, bean))
			v.Initialized()
		},
	)
}

func (container *di) destroyBean(beanName string, bean any) {
	dispatchLifecycle(bean,
		func(v DisposableWithContainer) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface DisposableWithContainer for %s(%T)", beanName, bean))
			v.Destroy(container)
		},
		func(v Disposable) {
			container.log.Debug(fmt.Sprintf("call lifecycle interface Disposable for %s(%T)", beanName, bean))
			v.Destroy()
		},
	)
}
