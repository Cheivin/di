---
layout: default
title: 循环依赖
nav_order: 8
parent: Bean 管理
---

# 循环依赖

di 的两阶段设计（**先全部实例化、再逐个注入依赖**）天然支持指针循环依赖。检测能力为 opt-in，默认关闭。

## 默认行为：支持循环依赖

默认情况下，`A ↔ B`、`A → B → C → A` 甚至 `A → A` 的指针循环引用都能正常注入：

```go
type A struct {
	B *B `aware:""`
}
type B struct {
	A *A `aware:""`
}

c := di.New()
c.Provide(A{})
c.Provide(B{})
c.Load()

a, _ := c.GetBean("a")
a.(*A).B.A == a.(*A) // true，循环引用闭环
```

### 原理

`Load` 时先为所有 bean 调用 `reflect.New` 创建指针对象存入 prototypeMap，再遍历注入依赖。注入 A 的 B 字段时，B 的指针已存在于 prototypeMap（虽然 B 的依赖可能尚未注入，但指针本身已可用），因此循环引用能闭环。

## 可选：开启严格检测

如果希望禁止循环依赖、保证依赖关系为有向无环图（DAG），可显式开启检测：

```go
c := di.New()
c.WithCircularCheck(true) // 开启循环依赖检测
c.Provide(A{})
c.Provide(B{})
c.Load() // panic: circular dependency: a -> b -> a
```

开启后，`Load` 时会对 aware 依赖图（含 `ProvideFunc` 工厂入参）做拓扑检测，发现环则 panic `ErrCircularDependency`。

### 检测范围

- **aware 标签字段依赖**：命名注入和接口类型的多候选匹配
- **ProvideFunc 工厂入参依赖**：工厂函数入参也算依赖
- `value`（配置项）不参与（纯数据，无环风险）

### 错误处理

panic 值是带 `Unwrap` 链的 error，可通过 `errors.Is` 判断：

```go
defer func() {
	if r := recover(); r != nil {
		err := r.(error)
		if errors.Is(err, di.ErrCircularDependency) {
			fmt.Println("循环依赖:", err)
			// 如：circular dependency: a -> b -> a
		}
	}
}()
c.Load()
```

### 常见环形态

| 形态 | 示例 |
|------|------|
| 直接环 | `A → B → A` |
| 间接环 | `A → B → C → A` |
| 自依赖 | `A → A` |

## 何时开启检测

| 场景 | 建议 |
|------|------|
| 通用业务（接受循环引用） | 默认关闭 |
| 希望依赖关系清晰、可测试 | 开启（保证 DAG） |
| 模块化架构（循环依赖视为设计问题） | 开启 |

## 如何解除循环依赖

若开启检测后遇到环，可：

- **重构**：把共享逻辑抽到第三个 bean，让 A、B 都依赖 C
- **延迟获取**：实现 `Injector` 接口拿到容器引用，运行期按需 `GetBean`
- **接口隔离**：把依赖收窄到小接口，打破类型层面的环

## 注意事项

- **仅检测 aware 与工厂入参**：`value` 配置不参与
- **找不到依赖不判环**：依赖指向不存在的 beanName 时检测跳过（注入期另行报 notfound）
- **RegisterBean 实例不参与**：手动注册的实例无 definition，不作为环的起点

> **变更说明**：循环依赖检测在 v0.3.1 引入，v0.3.1 至 v0.5.0 曾默认开启。v0.6.0 改为 opt-in（默认关闭）——因为 di 的两阶段设计本就支持指针循环依赖，默认检测属于回归 bug。

## 相关

- [生命周期](lifecycle) — `Load()` 的完整流程
- [构造函数注入](providefunc) — 工厂入参也参与检测
- [标签 aware](../tag/aware) — aware 标签的完整用法
