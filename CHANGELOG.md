# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [0.4.0] - 2026-08-08

### Breaking Changes

- **`Log` 接口 `Fatal` 签名变更**：`Fatal(string)` → `Fatal(error)`。所有自定义 `Log` 实现需同步修改签名；`recover()` 后可对 panic 值使用 `errors.Is` 判断具体错误类型（`ErrBean`/`ErrDefinition`/`ErrLoaded`/`ErrNotLoaded`/`ErrCircularDependency`）。
- **`DI` 接口新增方法**：`ProvideFunc(fn interface{})` 与 `WithBeanSelector(s BeanSelector)`。实现 `DI` 接口的外部类型需补充实现。
- **Go 最低版本提升至 1.25**：使用 `slices`/`maps`/`iter` 标准库特性，内部实现已泛型化。
- **`van` merge 行为修正**：`mergeStringMap` 在类型冲突（如 string vs map）时改为**新值覆盖旧值**，原为静默丢弃。

### 新增特性

- **构造函数注入 `ProvideFunc`**：注册工厂函数，容器按入参类型注入依赖，用返回值作为 bean。
- **接口歧义策略 `BeanSelector`**：内置 `LastRegistered`（默认，兼容旧行为）/`FirstRegistered`/`PrimaryFirst`/`ErrorOnAmbiguous` 四种策略，通过 `WithBeanSelector` 配置。
- **`Primary` 接口**：bean 实现 `IsPrimary() bool` 声明自己为首选实现。
- **slice/map 批量注入**：`[]T` 或 `map[string]T` 字段加 `aware:""` 标签，自动收集所有可赋值给 `T` 的 bean（slice 按注册顺序，map 以 beanName 为 key）。
- **循环依赖检测**：`Load()` 时拓扑检测，发现环直接 panic（`ErrCircularDependency`），错误信息包含环路径。
- **`ErrNotLoaded`**：`Serve()` 在未 `Load()` 时 panic 语义正确的错误。

### 修复

- **并发安全**：容器 `beanMap`/`beanDefinitionMap`/`prototypeMap`/`beanSort` 增加 `sync.RWMutex` 保护，支持运行期动态注册与并发读取（`GetBean`/`GetByType`/`GetByTypeAll`）。
- **`Fatal` 不再 `os.Exit`**：标准 logger 与 dio logger 的 `Fatal` 改为 panic，错误可被 `recover` 捕获，不再直接杀死进程。
- **`logger` 结构体指针接收者**：修复日志状态跨容器共享问题。
- **`in()` 函数删除**：原实现使用 `sort.SearchStrings` 做包含判断会原地排序入参（潜在 bug），已移除。

### 优化

- **`beanSort` 由 `container/list` 改为 `[]string`**：遍历使用 `slices.Backward` 等标准库，移除链表类型断言。
- **生命周期分发表驱动化**：`constructBean`/`preInitialize`/`afterPropertiesSet`/`initializedBean`/`destroyBean` 收敛为泛型 `dispatchLifecycle`，消除重复 type-switch。
- **`van` cast 增强**：支持 `fmt.Stringer` 转换、slice/array 元素逐个转换、未知类型 `fmt.Sprint` 兜底（不再静默丢值）；`time.Duration` 纯数字仍按毫秒兜底（行为兼容）。
- **`van` store 去反射**：`copyStringMap`/`mergeStringMap` 改原生 range，移除不必要反射。
- **全局容器懒初始化**：`di`/`dio` 包 import 不再创建容器，首次调用全局函数时才创建。
- **`Reset()`**：重置全局容器为未初始化状态，供测试隔离使用。

## [0.3.0] - 2025-06-01

### 新增特性

- 支持 `aware` 标签 `omitempty` 可选依赖。
- 支持 `BeanName` 接口自定义 bean 名称。
- 支持 `Injector` 接口自定义注入逻辑。
- `van` 配置管理器支持 `time.Duration` 类型转换（纯数字按毫秒）。

### 修复

- 根据类型查找 bean 时支持从手动注册的 bean 中查找。

## [0.2.0] - 2024-05-20

### 新增特性

- `ProvideOnProperty`/`ProvideNotOnProperty` 条件装配（dio 层）。
- `UnsafeMode` 不安全模式（私有字段注入）。
- 匿名结构体字段注入（`aware` 标签 + `Anonymous` 检查）。
