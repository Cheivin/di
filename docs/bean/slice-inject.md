---
layout: default
title: 批量注入
nav_order: 5
parent: Bean 管理
---

# 批量注入

{: .new-title}
> v0.4.0 新增
>
> slice / map 字段的 `aware` 标签批量收集能力在 v0.4.0 引入。

当一个接口有多个实现时，除了[单值注入 + 选择策略](selector)，还可以把它们**全部收集**到一个 slice 或 map 字段里。给 `[]T` 或 `map[string]T` 字段加上 `aware:""` 标签即可。

## slice 收集

`[]T` 字段加 `aware:""`，容器会收集所有可赋值给 `T` 的 bean，**按注册顺序**放入 slice：

```go
type Handler interface {
	Handle() string
}

type LogHandler struct{}
func (LogHandler) Handle() string { return "log" }

type AuthHandler struct{}
func (AuthHandler) Handle() string { return "auth" }

type Router struct {
	Handlers []Handler `aware:""` // 按注册顺序收集所有 Handler
}
```

容器在注入时会构造一个长度等于候选数的 slice，按 `beanSort` 顺序填入。元素类型 `T` 可以是接口，也可以是具体指针类型（如 `[]*UserRepo`）。

## map 收集

`map[string]T` 字段加 `aware:""`，容器以 beanName 为 key 收集所有实现：

```go
type Router struct {
	ByName map[string]Handler `aware:""` // beanName -> 实现
}
```

**map 的 key 必须是 `string`**，否则注册时 panic（错误包装 `ErrDefinition`）。这与 `beanName` 永远是字符串保持一致。

## 不走选择策略

slice/map 收集是**全量收集**，不经过 `BeanSelector`。这意味着即使一个接口有 100 个实现，也会全部进入 slice/map，不会因为歧义报错。这与单值 `aware:""` 字段的行为不同（单值注入多候选时由策略选定一个）。

| 字段类型 | 行为 | 多候选处理 |
|----------|------|-----------|
| `T` / `*T` / `Interface`（单值） | 注入一个 | `BeanSelector` 选定 |
| `[]T`（slice） | 全量收集 | 不走选择器 |
| `map[string]T`（map） | 全量收集 | 不走选择器 |

## 可选注入（omitempty）

slice/map 同样支持 `,omitempty`。当没有任何候选时：

- 加了 `aware:",omitempty"`：保持空 slice/map（非 nil 的空 slice），不报错
- 不加：保持空（slice 收集无候选时也会得到空 slice，不会 Fatal，因为空集合本身是合法状态）

```go
type Router struct {
	Handlers []Handler `aware:",omitempty"` // 无候选时为空 slice，不报错
}
```

注意：slice 注入即使无候选，注入的也是**非 nil 的空 slice**（`len == 0`），方便调用方直接 `range`。

## 示例

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

// Handler 是一个接口，有多个实现
type Handler interface {
	Handle() string
}

type LogHandler struct{}

func (LogHandler) Handle() string   { return "log" }
func (LogHandler) BeanName() string { return "logHandler" }

type AuthHandler struct{}

func (AuthHandler) Handle() string   { return "auth" }
func (AuthHandler) BeanName() string { return "authHandler" }

// Router 通过 slice 收集所有 Handler，通过 map 以 beanName 为 key 收集
type Router struct {
	Handlers []Handler          `aware:""` // 按注册顺序
	ByName   map[string]Handler `aware:""` // beanName -> 实现
}

func main() {
	di.Provide(LogHandler{})
	di.Provide(AuthHandler{})
	di.Provide(Router{})

	di.Load()

	router, _ := di.GetBean("router")
	r := router.(*Router)

	fmt.Printf("收集到 %d 个 handler（按注册顺序）:\n", len(r.Handlers))
	for _, h := range r.Handlers {
		fmt.Printf("  - %s\n", h.Handle())
	}
	// 输出：
	//   - log
	//   - auth

	fmt.Printf("\nmap 收集 %d 个（按 beanName）:\n", len(r.ByName))
	for name, h := range r.ByName {
		fmt.Printf("  %s -> %s\n", name, h.Handle())
	}
	// 输出：
	//   logHandler -> log
	//   authHandler -> auth
}
```

## 典型应用

- **中间件链**：`Middlewares []Middleware` 按注册顺序组装
- **路由表**：`Routes map[string]Handler` 以名称索引
- **插件系统**：`Plugins []Plugin` 收集所有已注册插件
- **多数据源**：`Repos []*UserRepo` 收集同类型的多个 bean

## 注意事项

- **map key 必须 string**：`map[int]T` 会在注册时 panic。
- **顺序保证**：slice 按 `beanSort`（注册顺序）；map 因 Go 内置 map 无序，遍历顺序不保证，但内容齐全。
- **元素类型**：可以是接口（`[]Handler`）或具体指针（`[]*DB`），只要是容器中 bean 可赋值的类型。
- **不参与歧义报错**：即便配置了 `ErrorOnAmbiguous` 策略，slice/map 收集也不会触发歧义错误。

## 相关

- [接口选择策略](selector) — 单值注入时的多实现选择
- [获取 bean](getbean) — `GetByTypeAll` 是运行期按类型取所有的等价方式
