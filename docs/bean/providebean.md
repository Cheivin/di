---
layout: default
title: 注册结构体
nav_order: 2
parent: Bean 管理
---

# 注册结构体

`Provide` / `ProvideNamedBean` 用于注册一个**结构体原型（值类型）**，容器在 `Load()` 时通过反射实例化为指针，并自动完成依赖注入与生命周期回调。

## 何时使用

- bean 可以无参构造（有默认零值即可），把创建工作交给容器
- 想用 `aware` / `value` 标签声明依赖，让容器自动注入

与 [RegisterBean](registerbean) 的区别：`Provide` 传入的是**值类型**（`UserService{}`），容器负责 `new`；`RegisterBean` 传入的是**已构造好的指针**。

## API

```go
// Provide 按类型推断 beanName 注册结构体原型
func Provide(prototype any) DI

// ProvideNamedBean 以指定名称注册结构体原型
func ProvideNamedBean(beanName string, prototype any) DI
```

## beanName 推断规则

与 `RegisterBean` 完全一致：

1. 若结构体实现了 `BeanName` 接口，使用 `BeanName()` 返回值（非空才生效）
2. 否则取类型名首字母小写（`GetBeanName`）：`UserService` → `userService`，`DB` → `dB`

`ProvideNamedBean` 的 `beanName` 参数优先级最高，会覆盖接口与默认推断。

## 示例

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct {
	Prefix string
}

type UserRepository struct {
	DB *DB `aware:"db"` // 按名注入 beanName 为 "db" 的实例
}

type UserService struct {
	Repo *UserRepository `aware:""` // 空名称：按类型推断 → "userRepository"
}

func main() {
	// 注册一个已实例化的 DB（名为 "db"）
	di.RegisterNamedBean("db", &DB{Prefix: "tbl_"})

	// 注册结构体原型，容器在 Load() 时反射 new 出指针
	di.Provide(UserService{})
	di.Provide(UserRepository{})

	di.Load()

	svc, _ := di.GetBean("userService")
	fmt.Printf("repo.DB.Prefix = %q\n", svc.(*UserService).Repo.DB.Prefix)
	// 输出: repo.DB.Prefix = "tbl_"
}
```

容器内部做的事情：

1. `reflect.New(UserService)` 创建 `*UserService`
2. 扫描 `aware` 标签字段，按名称/类型查找依赖并赋值
3. 扫描 `value` 标签字段，从配置项注入（带类型转换）
4. 依次触发生命周期回调（详见[生命周期](lifecycle)）

## 自定义 beanName

```go
type MQ struct {
	Topic string
}

// 方式一：实现 BeanName 接口
func (MQ) BeanName() string { return "kafka" }

// 方式二：注册时显式指定
di.ProvideNamedBean("rabbitmq", MQ{})
```

## 注意事项

- **传值，不是指针**：`Provide(UserService{})` 传值类型。容器内部会 `reflect.New` 出指针，原型值本身不会被使用（仅用于类型与字段定义）。
- **必须在 `Load()` 前注册**：`Provide` 在 `Load()` 之后调用会 Fatal（错误包装 `ErrLoaded`）。
- **重名会 Fatal**：beanName 与已注册的实例或定义冲突时 panic，错误包装 `ErrBean` / `ErrDefinition`。
- **单例**：`Provide` 注册的结构体在容器内是单例，所有依赖该类型的 bean 共享同一指针实例。需要每次新建实例请用 [NewBean](getbean)。

## 相关

- [注册实例](registerbean) — `RegisterBean` 注册已构造好的指针
- [构造函数注入](providefunc) — 需要构造逻辑时用工厂函数
- [获取 bean](getbean) — 取出容器实例化的 bean
