---
layout: default
title: 注册实例
nav_order: 1
parent: Bean 管理
---

# 注册实例

`RegisterBean` / `RegisterNamedBean` 用于将一个**已实例化**的 bean 注册到容器中。bean 必须是指针类型，容器会托管它的依赖注入与生命周期。

## 何时使用

- bean 的创建过程无法通过反射自动完成（例如需要传参、读配置、调用第三方构造函数）
- 想把外部已经构造好的对象（如 `*sql.DB`、`*redis.Client`）交给容器统一管理

## API

```go
// RegisterBean 按类型推断 beanName 注册
func RegisterBean(bean any) DI

// RegisterNamedBean 以指定名称注册
func RegisterNamedBean(name string, bean any) DI
```

两者区别仅在 beanName 的来源：`RegisterBean` 由容器推断名称，`RegisterNamedBean` 由调用方显式指定。

## beanName 推断规则

调用 `RegisterBean(bean)` 且未指定名称时，容器按以下优先级确定 beanName：

1. 若 bean 实现了 `BeanName` 接口，使用 `BeanName()` 的返回值（非空才生效）
2. 否则取类型名并首字母小写（`GetBeanName`）：`*DB` → `dB`，`*UserService` → `userService`

```go
type BeanName interface {
    BeanName() string
}
```

## bean 必须是指针

传入的 bean 必须是指针。传非指针（值类型、`nil`）会触发 `Fatal`（panic，错误包装 `ErrBean`）：

```go
di.RegisterBean(DB{})   // 错误：值类型，Fatal
di.RegisterBean(&DB{})  // 正确：指针
```

## 示例

```go
package main

import (
	"database/sql"
	"fmt"

	"github.com/cheivin/di"
)

type Cache struct {
	TTL int
}

// 实现 BeanName 接口自定义名称
func (c *Cache) BeanName() string { return "redisCache" }

func main() {
	// 1. 注册已有实例，按类型推断 beanName → "db"
	db, _ := sql.Open("sqlite3", ":memory:")
	di.RegisterBean(db)

	// 2. 注册已有实例，显式指定名称
	di.RegisterNamedBean("primaryCache", &Cache{TTL: 3600})

	// 3. beanName 由 BeanName() 接口决定 → "redisCache"
	di.RegisterBean(&Cache{TTL: 60})

	di.Load()

	primary, _ := di.GetBean("primaryCache")
	fmt.Printf("primary TTL = %d\n", primary.(*Cache).TTL)

	redis, _ := di.GetBean("redisCache")
	fmt.Printf("redis TTL = %d\n", redis.(*Cache).TTL)
}
```

## 注意事项

- **重名会 Fatal**：beanName 已存在时 panic，错误包装 `ErrBean`。同类型多实例务必用 `RegisterNamedBean` 指定不同名称。
- **注册时机**：应在 `Load()` 之前注册。`Load()` 之后注册的 bean 不会触发依赖注入与生命周期回调。
- **实例共享**：`RegisterBean` 注册的是单例，容器内所有依赖该类型的 bean 都会指向同一个实例。

## 相关

- [注册结构体](providebean) — `Provide` 注册类型原型，由容器实例化
- [获取 bean](getbean) — `GetBean` / `GetByType` 取出已注册的实例
- [生命周期](lifecycle) — `RegisterBean` 的实例同样走完整生命周期
