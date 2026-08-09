---
description: di 两阶段设计天然支持指针循环依赖；检测为 opt-in，默认关闭（v0.6.0 变更）。
---

# 循环依赖

di 的两阶段设计（**先全部实例化、再逐个注入依赖**）天然支持指针循环依赖。检测能力为 opt-in，默认关闭。

## 默认行为：支持循环依赖

默认情况下，`A ↔ B`、`A → B → C → A` 甚至 `A → A` 的指针循环引用都能正常注入：

```go
type A struct {
	B *B ` + "`aware:""`" + `
}
type B struct {
	A *A ` + "`aware:""`" + `
}

c := di.New()
c.Provide(A{})
c.Provide(B{})
c.Load()

a, _ := c.GetBean("a")
a.(*A).B.A == a.(*A) // true，循环引用闭环
```

### 原理

`Load` 时先为所有 bean 调用 `reflect.New` 创建指针对象存入 prototypeMap，再遍历注入依赖。注入 A 的 B 字段时，B 的指针已存在于 prototypeMap，循环引用能闭环。

## 可选：开启严格检测

```go
c := di.New()
c.WithCircularCheck(true) // 开启循环依赖检测
c.Provide(A{})
c.Provide(B{})
c.Load() // panic: circular dependency: a -> b -> a
```

### 错误处理

```go
defer func() {
	if r := recover(); r != nil {
		err := r.(error)
		if errors.Is(err, di.ErrCircularDependency) {
			fmt.Println("循环依赖:", err)
		}
	}
}()
c.Load()
```

## 何时开启检测

| 场景           | 建议         |
| ------------ | ---------- |
| 通用业务（接受循环引用） | 默认关闭       |
| 希望依赖关系清晰、可测试 | 开启（保证 DAG） |

> **变更说明**：循环依赖检测在 v0.3.1 引入，v0.3.1 至 v0.5.0 曾默认开启。v0.6.0 改为 opt-in（默认关闭）——因为 di 的两阶段设计本就支持指针循环依赖，默认检测属于回归 bug。
