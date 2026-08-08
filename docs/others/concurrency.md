---
layout: default
title: 并发安全
nav_order: 5
parent: 其他
---

# 并发安全

v0.4.0 起，di 容器内部用 `sync.RWMutex` 保护共享状态，支持运行期并发访问。

## 线程安全保证

| 操作 | 锁类型 | 说明 |
|------|--------|------|
| `GetBean` / `GetByType` / `GetByTypeAll` | 读锁 | 并发安全 |
| `RegisterBean` / `RegisterNamedBean` | 写锁 | 运行期可动态注册 |
| `Provide` / `ProvideNamedBean` | 写锁 | 仅 Load 前 |
| `NewBean` / `NewBeanByName` | 读锁（读定义）+ 无锁（实例化） | 实例化在锁外执行 |

## 运行期动态注册

`Load()` 之后仍可以注册新 bean，与并发读取不冲突：

```go
di.Load()

// 启动 HTTP 服务（多 goroutine 并发读）
go startServer()

// 运行期动态注册（写）
di.RegisterBean(&NewFeature{})
```

## 设计约束：回调在锁外执行

为保证不发生死锁，所有生命周期回调（`BeanConstruct`/`AfterPropertiesSet` 等）、`Injector.BeanInject`、日志调用都**在锁释放后执行**。这意味着：

- 回调函数内部可以安全地调用容器方法（`GetBean`/`RegisterBean`），不会死锁
- bean 实例引用一旦获取就稳定（`beanMap` 的值不替换，销毁时是 delete 而非覆盖）

## 不会死锁的原因

考虑这个场景：bean A 的 `BeanConstruct` 回调里调用了 `container.GetBean("B")`。

```
Load() → initializeBeans()
  └─ 锁内：取 A 的 definition 快照
  └─ 释放锁
  └─ 锁外：触发 A.BeanConstruct()
       └─ A 内部调 GetBean("B")
            └─ 加读锁，读 beanMap，释放读锁 ✅ 不死锁
```

如果回调在持锁时执行，`GetBean` 会重入同一把锁导致死锁——di 通过"锁内取数据、锁外跑回调"的纪律规避了这个问题。
