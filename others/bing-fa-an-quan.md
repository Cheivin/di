---
description: 容器线程安全保证与运行期动态注册（v0.4.0 新增）。
---

# 并发安全

v0.4.0 起，di 容器内部用 `sync.RWMutex` 保护共享状态。

## 线程安全保证

| 操作                      | 锁  | 说明       |
| ----------------------- | -- | -------- |
| `GetBean` / `GetByType` | 读锁 | 并发安全     |
| `RegisterBean`          | 写锁 | 运行期可动态注册 |
| `Provide`               | 写锁 | 仅 Load 前 |

## 运行期动态注册

```go
di.Load()
go startServer()              // 并发读
di.RegisterBean(&NewFeature{}) // 动态写，安全
```

## 设计约束

生命周期回调、`Injector.BeanInject`、日志调用都在**锁释放后**执行，因此回调内可安全调用容器方法，不会死锁。
