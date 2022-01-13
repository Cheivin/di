package di

import "reflect"

type (
	// BeanName 返回beanName
	BeanName interface {
		BeanName() string
	}

	// BeanConstruct Bean实例创建时
	BeanConstruct interface {
		BeanConstruct()
	}

	// BeanConstructWithContainer Bean实例创建时
	BeanConstructWithContainer interface {
		BeanConstruct(DI)
	}

	// PreInitialize Bean实例依赖注入前
	PreInitialize interface {
		PreInitialize()
	}

	// PreInitializeWithContainer Bean实例依赖注入前
	PreInitializeWithContainer interface {
		PreInitialize(DI)
	}

	InjectInfo struct {
		Bean        interface{}
		BeanName    string
		Type        reflect.Type
		IsPtr       bool // 是否为结构指针
		IsInterface bool // 是否为接口
		Anonymous   bool // 是否为匿名字段
		Omitempty   bool // 不存在依赖时则忽略注入
	}
	// Injector bean实例注入器
	Injector interface {
		BeanInject(di DI, info *InjectInfo)
	}

	// AfterPropertiesSet Bean实例注入完成
	AfterPropertiesSet interface {
		AfterPropertiesSet()
	}

	// AfterPropertiesSetWithContainer Bean实例注入完成
	AfterPropertiesSetWithContainer interface {
		AfterPropertiesSet(DI)
	}

	// Initialized 在Bean依赖注入完成后执行，可以理解为DI加载完成的通知事件。
	Initialized interface {
		Initialized()
	}

	// InitializedWithContainer 在Bean依赖注入完成后执行，可以理解为DI加载完成的通知事件。
	InitializedWithContainer interface {
		Initialized(DI)
	}

	// Disposable 在Bean注销时调用
	Disposable interface {
		Destroy()
	}
	// DisposableWithContainer 在Bean注销时调用
	DisposableWithContainer interface {
		Destroy(DI)
	}
)
