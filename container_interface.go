package di

import "context"

// DI 是依赖注入容器的核心接口。
// 通过链式方法注册 bean、设置配置，最后调用 Load 加载、Serve 运行。
type DI interface {
	// DebugMode 开启/关闭 debug 日志
	DebugMode(bool) DI

	// WithBeanSelector 设置接口多实现时的选择策略，默认 LastRegistered。
	WithBeanSelector(s BeanSelector) DI

	// WithCircularCheck 开启/关闭循环依赖检测，默认关闭（指针循环依赖可正常注入）
	WithCircularCheck(enable bool) DI

	// Log 设置容器的日志实现
	Log(log Log) DI

	// RegisterBean 注册一个已实例化的 bean（必须是指针），beanName 按类型推断
	RegisterBean(bean any) DI

	// RegisterNamedBean 以指定名称注册一个已实例化的 bean
	RegisterNamedBean(name string, bean any) DI

	// Provide 注册结构体原型（值类型），容器在 Load 时反射实例化为指针
	Provide(prototype any) DI

	// ProvideNamedBean 以指定名称注册结构体原型
	ProvideNamedBean(beanName string, prototype any) DI

	// ProvideFunc 注册工厂函数，容器按入参类型注入依赖，用返回值作为 bean。
	ProvideFunc(fn any) DI

	// GetBean 按名称获取 bean 实例
	GetBean(beanName string) (bean any, ok bool)

	// GetByType 按类型获取单个 bean（返回第一个匹配项）
	GetByType(beanType any) (bean any, ok bool)

	// GetByTypeAll 按类型获取所有匹配的 bean（实例），按注册顺序返回。
	// 注意：Serve 退出销毁 bean 后实例已从容器移除，基于实例的查询返回空；
	// 基于定义的查询（GetBeanNames/DescribeBean/GetBeanDependencies）不受影响
	GetByTypeAll(beanType any) (beans []BeanWithName)

	// GetBeanNames 返回所有已注册 bean 的名称（按注册顺序，含工厂 bean）。
	// 供管理/诊断场景使用（如统计容器 bean 数、描述容器内容）
	GetBeanNames() []string

	// HasBeanType 判断容器中是否已注册指定类型的 bean（实例/原型/工厂 bean 均可）。
	// beanType 可传值类型或指针类型，Load 前后均可用
	HasBeanType(beanType any) bool

	// DescribeBean 返回 bean 定义的只读描述（仅原型/工厂 bean 有定义）；不存在时返回 ok=false
	DescribeBean(beanName string) (desc BeanDescription, ok bool)

	// GetBeanDependencies 返回 bean 依赖的其他 bean 名称列表（命名 aware 注入，按名称排序）
	GetBeanDependencies(beanName string) (deps []string, ok bool)

	// NewBean 按类型创建新实例（非容器单例），走完整生命周期
	NewBean(beanType any) (bean any)

	// NewBeanByName 按名称创建新实例
	NewBeanByName(beanName string) (bean any)

	// UseValueStore 替换配置存储实现（默认 van）
	UseValueStore(v ValueStore) DI

	// Property 返回当前配置存储
	Property() ValueStore

	// SetDefaultProperty 设置默认配置项（低优先级，被 SetProperty 覆盖）
	SetDefaultProperty(key string, value any) DI

	// SetDefaultPropertyMap 批量设置默认配置项
	SetDefaultPropertyMap(properties map[string]any) DI

	// SetProperty 设置配置项（高优先级，覆盖 SetDefaultProperty）
	SetProperty(key string, value any) DI

	// SetPropertyMap 批量设置配置项
	SetPropertyMap(properties map[string]any) DI

	// AutoMigrateEnv 读取所有环境变量注入配置（key 中 _ 转为 .）
	AutoMigrateEnv() DI

	// GetProperty 获取配置项值
	GetProperty(key string) any

	// LoadProperties 将配置项（按 prefix 前缀）加载到结构体并返回新实例
	LoadProperties(prefix string, propertyType any) any

	// Load 加载容器：实例化、注入依赖、触发生命周期。重复调用会 panic
	Load()

	// Serve 阻塞等待 ctx 结束，然后倒序销毁所有 bean
	Serve(ctx context.Context)

	// Context 返回容器的 context（Serve 时设置）
	Context() context.Context
}
