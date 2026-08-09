---
description: 从容器中获取被托管的实例，不论是手动注册的实例还是结构体
---

# 获取bean

容器提供了按名称、按类型获取 bean，以及每次新建实例（非单例）的多种方式。

## API 一览

| 方法                       | 返回               | 说明                     |
| ------------------------ | ---------------- | ---------------------- |
| `GetBean(name)`          | `(any, bool)`    | 按名获取单例                 |
| `GetByType(beanType)`    | `(any, bool)`    | 按类型获取单个单例（多个时取注册顺序第一个） |
| `GetByTypeAll(beanType)` | `[]BeanWithName` | 按类型获取所有单例（按注册顺序）       |
| `NewBean(beanType)`      | `any`            | 按类型每次新建实例（非单例）         |
| `NewBeanByName(name)`    | `any`            | 按名每次新建实例（非单例）          |

## 根据名称获取（GetBean）

```go
func GetBean(beanName string) (bean any, ok bool)
```

最直接的获取方式，`ok` 为 false 表示该名称未注册。

```go
svc, ok := di.GetBean("userService")
if !ok {
    panic("userService not found")
}
u := svc.(*UserService)
```

## 根据类型获取（GetByType）

```go
func GetByType(beanType any) (bean any, ok bool)
```

按类型查找所有可赋值的 bean，返回一个。当存在多个候选时，取注册顺序中的第一个。`beanType` 需要携带类型信息：

```go
// 1. 传值：最直观
bean, ok := di.GetByType(&UserService{})

// 2. 传类型化 nil 指针：不需要构造实例
var p *UserService
bean, ok := di.GetByType(p)

// 3. 传接口类型：找到所有实现该接口的 bean
var sender Sender
bean, ok := di.GetByType(&sender)
```

注意：直接传 `nil`（无类型）会返回空结果，因为 `reflect.TypeOf(nil) == nil`。

## GetByTypeAll 获取所有

```go
func GetByTypeAll(beanType any) (beans []BeanWithName)
```

返回所有可赋值给 `beanType` 的 bean，**按注册顺序排列**。每个元素包含 beanName 与实例：

```go
type BeanWithName struct {
    Name string
    Bean any
}

handlers := di.GetByTypeAll((*Handler)(nil))
for _, h := range handlers {
    fmt.Printf("%s -> %T\n", h.Name, h.Bean)
}
```

用于需要遍历所有实现（如插件、中间件、事件处理器）的场景，也可以在字段上用 slice 批量注入让容器自动收集。

## 管理诊断 API（v0.6.2 新增）

除获取 bean 外，容器还提供一组只读的管理/诊断 API，用于统计 bean 数、描述 bean 定义、排查依赖关系：

| 方法                          | 返回                        | 说明                          |
| --------------------------- | ------------------------- | --------------------------- |
| `GetBeanNames()`            | `[]string`                | 所有 bean 名称（按注册顺序，含工厂 bean）  |
| `HasBeanType(beanType)`     | `bool`                    | 是否已注册指定类型的 bean（实例/原型/工厂均可） |
| `DescribeBean(name)`        | `(BeanDescription, bool)` | bean 定义的只读描述                |
| `GetBeanDependencies(name)` | `([]string, bool)`        | 依赖的其他 bean 名称列表（按名称排序）      |

### GetBeanNames 统计 bean

```go
names := di.GetBeanNames()
fmt.Println(len(names)) // 容器内 bean 总数
```

### HasBeanType 判断类型是否已注册

```go
if di.HasBeanType(&UserService{}) {
    // 该类型已有 bean（实例/原型/工厂均可命中）
}
```

`beanType` 的传参规则与 `GetByType` 一致（值类型或类型化 nil 指针）。

### DescribeBean / GetBeanDependencies 描述定义

```go
desc, ok := di.DescribeBean("userService")
if ok {
    fmt.Println(desc.Name, desc.Type, desc.Factory)
    for _, dep := range desc.Dependencies {
        fmt.Println(dep.Field, "->", dep.Name)
    }
    for _, v := range desc.Values {
        fmt.Println(v.Field, "=", v.Name)
    }
}

deps, ok := di.GetBeanDependencies("userService") // ["db"]
```

`BeanDescription` 包含 bean 名称、类型、是否工厂模式，以及 aware 依赖（`Dependencies`）与 value 配置注入（`Values`）列表；`Dependency` 含字段名、注入的 bean 名或配置项 key、类型与是否可选（`omitempty`）。

注意事项：

* **定义信息仅原型（`Provide`）与工厂（`ProvideFunc`）bean 有**；直接注册的实例（`RegisterBean`）不经过定义解析，`DescribeBean` 返回 `ok=false`。
* `GetBeanDependencies` 只包含命名注入（`aware` 指定名称或按类型推断出名称的）；slice/map 按类型收集的注入不包含。
* **基于实例的查询（`GetBean`/`GetByType`/`GetByTypeAll`）在 `Serve` 退出、bean 销毁后返回空**；基于定义的查询（`GetBeanNames`/`DescribeBean`/`GetBeanDependencies`）不受影响，可放心用于停机后的诊断。

## NewBean / NewBeanByName 每次新建

```go
func NewBean(beanType any) (bean any)
func NewBeanByName(beanName string) (bean any)
```

与 `GetBean`/`GetByType` 返回容器内单例不同，`NewBean` 每次调用都**创建新实例**，并走完整生命周期（`BeanConstruct` → 注入 → `AfterPropertiesSet` → `Initialized`）。当该实例被 GC 回收时，容器会通过 `runtime.SetFinalizer` 触发 `Destroy` 回调。

```go
// 按类型新建
req1 := di.NewBean(&RequestContext{}).(*RequestContext)
req2 := di.NewBean(&RequestContext{}).(*RequestContext)
// req1 != req2，是两个独立实例

// 按名新建（必须已用 Provide/ProvideNamedBean 注册过定义）
scope := di.NewBeanByName("requestScope")
```

`NewBeanByName` 找不到定义时会 panic（错误包装 `ErrDefinition`）；`NewBean` 找不到已注册定义时则会现场构造一个新的 definition 并实例化。

## 注意事项

* **GetBean/GetByType 是单例**：多次调用返回同一指针。
* **类型 nil**：`GetByType(nil)` 返回空，传参必须携带类型信息。
* **NewBean 的销毁时机**：依赖 GC，不保证立即触发；不要在 `Destroy` 里做时间敏感的清理。
* **并发安全**：所有获取方法在 `sync.RWMutex` 保护下可并发调用，也支持 `Load()` 后动态注册的 bean 立即可见。
