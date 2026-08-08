package di

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"runtime"
	"slices"
	"sync"

	"github.com/cheivin/di/van"
)

type (
	di struct {
		log               Log
		beanDefinitionMap map[string]definition  // Name:bean定义
		prototypeMap      map[string]any // Name:初始化的bean
		beanMap           map[string]any // Name:bean实例
		loaded            bool
		unsafe            bool
		valueStore        ValueStore
		beanSort          []string // 注册顺序（beanName）
		ctx               context.Context
		mu                sync.RWMutex // 保护 beanDefinitionMap/prototypeMap/beanMap/beanSort
		selector          BeanSelector
	}
)

func (container *di) Context() context.Context {
	return container.ctx
}

// withRLock 在读锁保护下执行 fn。
// 仅用于纯读路径；涉及锁外回调（生命周期/Injector）的路径不得使用。
// 注意：Go 不允许方法带类型参数，故为包级泛型函数。
func withRLock[T any](container *di, fn func() T) T {
	container.mu.RLock()
	defer container.mu.RUnlock()
	return fn()
}

// withLock 在写锁保护下执行 fn。
func withLock(container *di, fn func()) {
	container.mu.Lock()
	defer container.mu.Unlock()
	fn()
}

var (
	ErrBean       = errors.New("error bean")
	ErrDefinition = errors.New("error definition")
	ErrLoaded     = errors.New("di loaded")
	ErrNotLoaded  = errors.New("di not loaded")
)

// New 创建一个新的 DI 容器实例。
// 返回 *di（同时实现 DI 接口），供需要独立容器实例的场景使用。
// 全局函数（di.RegisterBean 等）操作一个独立的全局容器，见 [container]。
func New() *di {
	return &di{
		log:               stdLogger(),
		beanDefinitionMap: map[string]definition{},
		prototypeMap:      map[string]any{},
		beanMap:           map[string]any{},
		valueStore:        van.New(),
		beanSort:          []string{},
		ctx:               context.Background(),
		selector:          LastRegistered{},
	}
}

// UnsafeMode 开启不安全模式，允许通过 unsafe.Pointer 注入未导出（私有）字段。
// 开启后影响所有 value/aware 注入，容器会打印 warn 日志。
func (container *di) UnsafeMode(open bool) DI {
	container.unsafe = open
	container.log.Warn("Unsafe mode enabled!")
	return container
}

// parseBeanType 解析 bean 类型并推断 beanName。
// beanName 优先级：显式传入 > BeanName 接口返回值 > 类型名首字母小写（GetBeanName）。
func (container *di) parseBeanType(beanType any) (prototype reflect.Type, beanName string) {
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

// DebugMode 开启/关闭 debug 日志。
func (container *di) DebugMode(enable bool) DI {
	container.log.DebugMode(enable)
	return container
}

// WithBeanSelector 设置接口多实现时的选择策略。
// 传入 nil 则恢复默认的 LastRegistered。必须在 Load 前调用。
func (container *di) WithBeanSelector(s BeanSelector) DI {
	if s == nil {
		s = LastRegistered{}
	}
	container.selector = s
	return container
}

// Log 设置容器的日志实现。
func (container *di) Log(log Log) DI {
	container.log = log
	return container
}

// RegisterBean 注册一个已生成的bean，根据bean类型生成beanName
func (container *di) RegisterBean(bean any) DI {
	return container.RegisterNamedBean("", bean)
}

// RegisterNamedBean 以指定名称注册一个bean
func (container *di) RegisterNamedBean(beanName string, bean any) DI {
	if !IsPtr(bean) {
		container.log.Fatal(fmt.Errorf("%w: bean must be a pointer", ErrBean))
		return container
	}
	if beanName == "" {
		_, beanName = container.parseBeanType(bean)
	}
	container.mu.Lock()
	defer container.mu.Unlock()
	if _, exist := container.beanMap[beanName]; exist {
		container.log.Fatal(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
		return container
	}
	container.beanMap[beanName] = bean
	// 加入队列
	container.beanSort = append(container.beanSort, beanName)
	container.log.Info(fmt.Sprintf("register bean with name: %s", beanName))
	return container
}

// Provide 注册结构体原型（值类型），容器在 Load 时反射实例化为指针并注入依赖。
// beanName 按 [parseBeanType] 规则推断。
func (container *di) Provide(prototype any) DI {
	container.ProvideNamedBean("", prototype)
	return container
}

// ProvideFunc 注册工厂函数，容器按入参类型注入依赖，用返回值作为 bean。
// fn 必须是 func(...) (...)，且只有一个返回值（指针类型）。
func (container *di) ProvideFunc(fn any) DI {
	if container.loaded {
		container.log.Fatal(fmt.Errorf("%w", ErrLoaded))
		return container
	}
	fv := reflect.ValueOf(fn)
	if fv.Kind() != reflect.Func {
		container.log.Fatal(fmt.Errorf("%w: ProvideFunc expects a function, got %T", ErrBean, fn))
		return container
	}
	ft := fv.Type()
	if ft.NumOut() != 1 {
		container.log.Fatal(fmt.Errorf("%w: ProvideFunc must return exactly one value, got %d", ErrBean, ft.NumOut()))
		return container
	}
	returnType := ft.Out(0)
	if returnType.Kind() != reflect.Ptr {
		container.log.Fatal(fmt.Errorf("%w: ProvideFunc must return a pointer, got %s", ErrBean, returnType.String()))
		return container
	}
	// beanName：优先用返回类型实现 BeanName 接口的值，否则按类型推断
	beanName := ""
	tmpBean := reflect.New(returnType.Elem()).Interface()
	if bn, ok := tmpBean.(BeanName); ok {
		beanName = bn.BeanName()
	}
	if beanName == "" {
		beanName = GetBeanName(returnType)
	}
	// 入参类型列表
	args := make([]reflect.Type, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		args[i] = ft.In(i)
	}
	def := definition{
		Name:        beanName,
		Type:        returnType,
		awareMap:    map[string]aware{},
		valueMap:    map[string]aware{},
		factory:     fv,
		factoryArgs: args,
	}
	container.mu.Lock()
	defer container.mu.Unlock()
	if _, exist := container.beanMap[beanName]; exist {
		container.log.Fatal(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
		return container
	}
	if existDef, exist := container.beanDefinitionMap[beanName]; exist {
		container.log.Fatal(fmt.Errorf("%w: bean %s already defined by %s", ErrDefinition, beanName, existDef.Type.String()))
		return container
	}
	container.beanDefinitionMap[beanName] = def
	container.beanSort = append(container.beanSort, beanName)
	container.log.Info(fmt.Sprintf("provide bean(factory) with name: %s", beanName))
	return container
}

// ProvideNamedBean 以指定名称注册结构体原型。
// beanName 为空时按 [parseBeanType] 推断。Load 后调用会 Fatal。
func (container *di) ProvideNamedBean(beanName string, beanType any) DI {
	if container.loaded {
		container.log.Fatal(fmt.Errorf("%w", ErrLoaded))
		return container
	}
	var prototype reflect.Type
	if beanName == "" {
		prototype, beanName = container.parseBeanType(beanType)
	} else {
		prototype, _ = container.parseBeanType(beanType)
	}
	// newDefinition 含反射与日志，在锁外执行避免长持有
	def := container.newDefinition(beanName, prototype)

	container.mu.Lock()
	defer container.mu.Unlock()
	// 检查bean重复
	if _, exist := container.beanMap[beanName]; exist {
		container.log.Fatal(fmt.Errorf("%w: bean %s already exists", ErrBean, beanName))
		return container
	}
	// 检查beanDefinition重复
	if existDefinition, exist := container.beanDefinitionMap[beanName]; exist {
		container.log.Fatal(fmt.Errorf("%w: bean %s already defined by %s", ErrDefinition, beanName, existDefinition.Type.String()))
		return container
	}
	container.beanDefinitionMap[beanName] = def
	// 加入队列
	container.beanSort = append(container.beanSort, beanName)
	container.log.Info(fmt.Sprintf("provide bean with name: %s", beanName))
	return container
}

// GetBean 按名称获取 bean 实例。线程安全（读锁）。
func (container *di) GetBean(beanName string) (bean any, ok bool) {
	container.mu.RLock()
	defer container.mu.RUnlock()
	bean, ok = container.beanMap[beanName]
	return
}

// getAllByType 按类型查找 bean。beanType 接受值类型或指针类型（如 T{} 或 (*T)(nil)）。
// limitOne 为 true 时找到第一个即返回（GetByType 使用）。
// 线程安全（读锁）。
func (container *di) getAllByType(beanType any, limitOne bool) (beans []BeanWithName) {
	// 空/nil 参数直接返回空
	t := reflect.TypeOf(beanType)
	if t == nil {
		return
	}
	var typeValue reflect.Type
	if t.Kind() == reflect.Ptr {
		typeValue = t.Elem()
		if typeValue.Kind() == reflect.Struct {
			typeValue = reflect.PtrTo(typeValue)
		}
	} else {
		typeValue = reflect.PtrTo(t)
	}
	container.mu.RLock()
	defer container.mu.RUnlock()
	for name, bean := range container.beanMap {
		if reflect.TypeOf(bean).AssignableTo(typeValue) {
			beans = append(beans, BeanWithName{
				Name: name,
				Bean: bean,
			})
			if limitOne {
				return
			}
		}
	}
	return
}

// GetByType 按类型获取单个 bean（返回第一个匹配项）。
// beanType 可传值类型或指针类型，如 di.GetByType(&UserService{}) 或 di.GetByType((*Service)(nil))。
func (container *di) GetByType(beanType any) (any, bool) {
	beans := container.getAllByType(beanType, true)
	if len(beans) == 0 {
		return nil, false
	} else {
		return beans[0].Bean, true
	}
}

// GetByTypeAll 按类型获取所有匹配的 bean（含名称），按注册顺序返回。
func (container *di) GetByTypeAll(beanType any) (beans []BeanWithName) {
	return container.getAllByType(beanType, false)
}

// NewBean 按类型创建一个新的 bean 实例（非容器单例）。
// 新实例会走完整生命周期（BeanConstruct → 注入 → AfterPropertiesSet → Initialized），
// 并通过 runtime.SetFinalizer 在 GC 时触发 Destroy。
// 已有 definition 时复用，否则临时构造。
func (container *di) NewBean(beanType any) (bean any) {
	prototype, beanName := container.parseBeanType(beanType)
	// 检查beanDefinition是否存在
	exist := withRLock(container, func() bool {
		_, exist := container.beanDefinitionMap[beanName]
		return exist
	})
	if !exist {
		return container.newBean(container.newDefinition(beanName, prototype))
	}
	return container.NewBeanByName(beanName)
}

// NewBeanByName 按名称创建新 bean 实例。definition 不存在时 panic（ErrDefinition）。
func (container *di) NewBeanByName(beanName string) (bean any) {
	container.mu.RLock()
	def, ok := container.beanDefinitionMap[beanName]
	container.mu.RUnlock()
	if !ok {
		panic(fmt.Errorf("%w: %s notfound", ErrDefinition, beanName))
	}
	return container.newBean(def)
}

// newBean 按定义创建并完整初始化一个 bean 实例（实例化 → 构造 → 注入 → 初始化 → 注册 finalizer）。
// 注：finalizer 由 GC 在任意 goroutine 触发调用 destroyBean，destroyBean 仅调用 Destroy 回调不碰 beanMap，
// 因此与 destroyBeans（锁内 delete）不会产生数据竞争。
func (container *di) newBean(def definition) (bean any) {
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
	runtime.SetFinalizer(bean, func(bean any) {
		container.destroyBean(def.Name, bean)
	})
	return
}

// Load 加载容器：实例化所有 bean、注入依赖、触发生命周期回调。
// 首先做循环依赖检测（失败则 panic ErrCircularDependency，且 loaded 置回 false 允许 recover 后重试）。
// Load 后再次调用会 panic ErrLoaded。
func (container *di) Load() {
	if container.loaded {
		panic(ErrLoaded)
	}

	container.loaded = true
	// 注入前先做循环依赖检测，避免运行期未定义行为。
	// 检测失败时还原 loaded，允许调用方 recover 后修正依赖并重新 Load。
	if err := container.checkCircularDependency(); err != nil {
		container.loaded = false
		panic(err)
	}
	container.initializeBeans()
	container.processBeans()
	container.initialized()

}

// Serve 阻塞等待 ctx 结束，然后倒序销毁所有 bean（触发 Destroy 回调）。
// 必须在 Load 之后调用，否则 panic ErrNotLoaded。
// 通常配合 signal.NotifyContext 监听 SIGINT/SIGTERM 使用。
func (container *di) Serve(ctx context.Context) {
	if !container.loaded {
		panic(ErrNotLoaded)
	}
	var cancel context.CancelFunc
	container.ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	<-ctx.Done()
	container.destroyBeans()
}

// initializeBeans 初始化bean对象
func (container *di) initializeBeans() {
	// 锁内收集 definition 快照，释放锁后再实例化
	// （工厂 bean 实例化时会反向调 findBeanByName/findBeanByType，持锁会死锁）
	container.mu.Lock()
	snapshot := slices.Collect(maps.Values(container.beanDefinitionMap))
	container.mu.Unlock()
	// 创建类型的指针对象（instanceBean 含工厂调用/value 注入/日志，必须在锁外）
	prototypes := make(map[string]any, len(snapshot))
	for _, def := range snapshot {
		prototypes[def.Name] = container.instanceBean(def)
	}
	container.mu.Lock()
	maps.Copy(container.prototypeMap, prototypes)
	container.mu.Unlock()
	// 根据排序遍历触发BeanConstruct方法（回调可能反向访问容器，必须在锁外）
	for _, beanName := range container.beanSort {
		container.mu.RLock()
		prototype, ok := container.prototypeMap[beanName]
		container.mu.RUnlock()
		if ok {
			container.constructBean(beanName, prototype)
		}
	}
}

// processBeans 注入依赖
func (container *di) processBeans() {
	for _, beanName := range container.beanSort {
		container.mu.RLock()
		prototype, ok := container.prototypeMap[beanName]
		def := container.beanDefinitionMap[beanName]
		container.mu.RUnlock()
		if !ok {
			continue
		}
		// 加载为bean
		container.log.Info(fmt.Sprintf("initialize bean %s(%T)", def.Name, prototype))
		// processBean 含 PreInitialize/wireBean/AfterPropertiesSet 回调，必须在锁外执行
		bean := container.processBean(prototype, def)
		container.mu.Lock()
		container.beanMap[beanName] = bean
		container.mu.Unlock()
	}
}

// initialized 容器初始化完成
func (container *di) initialized() {
	for _, beanName := range container.beanSort {
		container.mu.RLock()
		bean := container.beanMap[beanName]
		container.mu.RUnlock()
		// 回调在锁外
		container.initializedBean(beanName, bean)
	}
}

// destroyBeans 按注册倒序销毁 bean：锁内从 beanMap 移除，锁外触发 Destroy 回调。
func (container *di) destroyBeans() {
	// 倒序销毁bean
	for _, beanName := range slices.Backward(container.beanSort) {
		container.mu.Lock()
		bean, ok := container.beanMap[beanName]
		if ok {
			delete(container.beanMap, beanName)
		}
		container.mu.Unlock()
		if ok {
			// 回调在锁外
			container.destroyBean(beanName, bean)
		}
	}
}
