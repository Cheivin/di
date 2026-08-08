---
layout: default
title: 循环依赖检测
nav_order: 8
parent: Bean 管理
---

# 循环依赖检测

{: .new-title}
> v0.4.0 新增
>
> 启动期循环依赖检测在 v0.4.0 引入，让依赖环在 `Load()` 阶段即暴露，避免运行期未定义行为。

## 它解决什么问题

A 依赖 B、B 又依赖 A，构成环。在没有检测时，这种依赖会导致注入得到 nil 或死锁，问题难以定位。`di` 在 `Load()` 时对依赖图做拓扑遍历（DFS），一旦发现环立即 panic，错误信息包含完整的环路径。

## 检测范围

`Load()` 时自动检测以下两类依赖构成的环：

- **`aware` 标签字段依赖**：包括命名注入和接口类型的多候选匹配
- **`ProvideFunc` 工厂入参依赖**：工厂函数的入参也算依赖

`value`（配置项）不参与检测。无名称的接口依赖会按类型匹配所有实现并纳入检测。

## 检测时机

在 `Load()` 的最开头、所有实例化之前执行。一旦发现环，`Load()` 直接 panic，容器不会进入半初始化状态（`loaded` 标志会被回退）。

## 错误信息

panic 值是 error 类型，错误信息形如：

```
circular dependency: cycleA -> cycleB -> cycleA
```

包含从环的入口到回到自身的完整路径。底层使用 `ErrCircularDependency` 作为哨兵错误，可通过 `errors.Is` 判断。

## 判断错误类型

`Load()` 的 panic 可被 `recover` 捕获，再用 `errors.Is` 判断是否为循环依赖：

```go
var ErrCircularDependency = errors.New("circular dependency")

defer func() {
    if r := recover(); r != nil {
        err, ok := r.(error)
        if ok && errors.Is(err, di.ErrCircularDependency) {
            // 处理循环依赖
            fmt.Println("检测到循环依赖:", err)
        } else {
            // 重新抛出其他错误
            panic(r)
        }
    }
}()
di.Load()
```

## 示例

```go
package main

import (
	"errors"
	"fmt"

	"github.com/cheivin/di"
)

// A → B → A 构成循环
type A struct {
	B *B `aware:""`
}

type B struct {
	A *A `aware:""`
}

// 正常的依赖链 C → D（无环）
type C struct {
	D *D `aware:""`
}

type D struct{}

func main() {
	fmt.Println("===== 正常依赖（无环）=====")
	c1 := di.New()
	c1.Provide(C{})
	c1.Provide(D{})
	c1.Load()
	fmt.Println("加载成功")

	fmt.Println("\n===== 循环依赖（A → B → A）=====")
	c2 := di.New()
	c2.Provide(A{})
	c2.Provide(B{})

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			if errors.Is(err, di.ErrCircularDependency) {
				fmt.Printf("检测到循环依赖: %v\n", err)
				// 输出：检测到循环依赖: circular dependency: a -> b -> a
			} else {
				panic(r) // 非循环依赖错误，重新抛出
			}
		}
	}()
	c2.Load()
}
```

输出：

```
===== 正常依赖（无环）=====
加载成功

===== 循环依赖（A → B → A）=====
检测到循环依赖: circular dependency: a -> b -> a
```

## 常见环形态

检测能识别各种环：

| 形态 | 示例 |
|------|------|
| 直接环 | `A → B → A` |
| 间接环 | `A → B → C → A` |
| 自依赖 | `A → A`（A 的字段依赖自己）|

## 如何解除循环依赖

- **重构**：把共享逻辑抽到第三个 bean，让 A、B 都依赖 C，而不是互相依赖。
- **延迟获取**：实现 `Injector` 接口在注入时拿到容器引用，运行期再按需 `GetBean`，而不是声明期就依赖。
- **接口隔离**：把 A 对 B 的依赖收窄到一个小接口，打破类型层面的环。
- **配置注入**：部分场景下用 `value` 配置替代 bean 依赖可绕开环。

## 注意事项

- **仅检测 aware 与工厂入参**：`value` 配置不参与，因为配置项是纯数据、无环风险。
- **找不到依赖不判环**：若依赖指向不存在的 beanName，检测阶段会跳过（注入期会另行报 notfound），不会误判为环。
- **RegisterBean 实例不参与**：手动注册的实例无 definition，不作为环的起点；但其类型若被其他 bean 依赖，仍会作为依赖目标纳入检测。

## 相关

- [生命周期](lifecycle) — `Load()` 的后续阶段
- [构造函数注入](providefunc) — 工厂入参也参与检测
- [标签 aware](../tag/aware) — aware 标签的完整用法
