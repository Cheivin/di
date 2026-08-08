---
layout: default
title: 获取 bean
nav_order: 4
parent: Bean 管理
---

# 获取 bean

容器提供了按名称、按类型获取 bean，以及每次新建实例（非单例）的多种方式。

## API 一览

| 方法 | 返回 | 说明 |
|------|------|------|
| `GetBean(name)` | `(any, bool)` | 按名获取单例 |
| `GetByType(beanType)` | `(any, bool)` | 按类型获取单个单例（多个时取一个） |
| `GetByTypeAll(beanType)` | `[]BeanWithName` | 按类型获取所有单例 |
| `NewBean(beanType)` | `any` | 按类型每次新建实例（非单例） |
| `NewBeanByName(name)` | `any` | 按名每次新建实例（非单例） |

## GetBean 按名获取

```go
func GetBean(beanName string) (bean any, ok bool)
```

最直接的获取方式。`ok` 为 false 表示该名称未注册。

```go
svc, ok := di.GetBean("userService")
if !ok {
    panic("userService not found")
}
u := svc.(*UserService)
```

## GetByType 按类型获取

```go
func GetByType(beanType any) (bean any, ok bool)
```

按类型查找所有可赋值的 bean，返回一个。当存在多个候选时，取注册顺序中的第一个（非接口歧义场景，单值注入的歧义处理见 [接口选择策略](selector)）。

### 如何传类型参数

`beanType` 需要携带类型信息，有以下几种写法：

```go
// 1. 传值：最直观
bean, ok := di.GetByType(&UserService{}) // ok

// 2. 传类型化 nil 指针：不需要构造实例
var p *UserService
bean, ok := di.GetByType(p) // 注意 p 必须是带类型的 nil，不能用裸 nil

// 3. 传接口类型
var sender Sender
bean, ok := di.GetByType(&sender) // 容器内部会按 reflect.TypeOf 处理
```

注意：直接传 `nil`（无类型）会返回空结果，因为 `reflect.TypeOf(nil) == nil`。

### 接口类型

`GetByType` 同样适用于接口：会找到所有实现了该接口的 bean。

```go
type Sender interface{ Send(string) }

// 假设注册了 EmailSender / SmsSender，都实现 Sender
bean, ok := di.GetByType((*Sender)(nil)) // 取一个实现
```

## GetByTypeAll 获取所有

```go
func GetByTypeAll(beanType any) (beans []BeanWithName)
```

返回所有可赋值给 `beanType` 的 bean，按注册顺序排列。每个元素包含 beanName 与实例：

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

用于需要遍历所有实现（如插件、中间件、事件处理器）的场景。也可以直接在字段上用 [slice 批量注入](slice-inject) 让容器自动收集。

## NewBean / NewBeanByName 每次新建

```go
func NewBean(beanType any) (bean any)
func NewBeanByName(beanName string) (bean any)
```

与 `GetBean`/`GetByType` 返回容器内单例不同，`NewBean` 每次调用都**创建新实例**，并走完整生命周期（`BeanConstruct` → 注入 → `AfterPropertiesSet` → `Initialized`）。当该实例被 GC 回收时，容器会通过 `runtime.SetFinalizer` 触发 `Destroy` 回调。

适合需要独立状态的临时对象（如每次请求一个上下文 bean）。

```go
// 按类型新建
req1 := di.NewBean(&RequestContext{}).(*RequestContext)
req2 := di.NewBean(&RequestContext{}).(*RequestContext)
// req1 != req2，是两个独立实例

// 按名新建（必须已用 Provide/ProvideNamedBean 注册过定义）
scope := di.NewBeanByName("requestScope")
```

`NewBeanByName` 找不到定义时会 panic（错误包装 `ErrDefinition`）；`NewBean` 找不到已注册定义时则会现场构造一个新的 definition 并实例化。

## 示例

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct{}

type Service struct {
	DB *DB `aware:""`
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(Service{})
	di.Load()

	// 按名获取
	svc, ok := di.GetBean("service")
	fmt.Println(ok, svc.(*Service).DB != nil) // true true

	// 按类型获取
	bean, ok := di.GetByType(&Service{})
	fmt.Println(ok, bean.(*Service) != nil) // true true

	// 获取所有 DB 类型实例
	all := di.GetByTypeAll(&DB{})
	fmt.Println(len(all)) // 1

	// 每次新建（非单例）
	a := di.NewBean(&Service{})
	b := di.NewBean(&Service{})
	fmt.Println(a != b) // true，独立实例
}
```

## 注意事项

- **GetBean/GetByType 是单例**：多次调用返回同一指针。
- **类型 nil**：`GetByType(nil)` 返回空，传参必须携带类型信息。
- **NewBean 的销毁时机**：依赖 GC，不保证立即触发；不要在 `Destroy` 里做时间敏感的清理。
- **并发安全**：所有获取方法在 `sync.RWMutex` 保护下可并发调用，也支持 `Load()` 后动态注册的 bean 立即可见。

## 相关

- [注册实例](registerbean) / [注册结构体](providebean) — 先注册才能获取
- [批量注入](slice-inject) — 用字段直接收集所有同类型 bean
- [生命周期](lifecycle) — `NewBean` 的实例也走完整生命周期
