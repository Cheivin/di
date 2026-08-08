---
description: slice/map 字段加 aware 标签自动收集同类型的所有 bean（v0.4.0 新增）。
---

# 批量注入（slice / map）

`[]T` 或 `map[string]T` 字段加 `aware:""` 标签，自动收集所有可赋值给 `T` 的 bean。

## slice 收集

```go
type Handler interface {
	Handle() string
}

type Router struct {
	Handlers []Handler `aware:""` // 收集所有 Handler 实现，按注册顺序
}

di.Provide(LogHandler{})
di.Provide(AuthHandler{})
di.Provide(Router{})
di.Load()
// Router.Handlers = [LogHandler, AuthHandler]
```

## map 收集

```go
type Router struct {
	ByName map[string]Handler `aware:""` // beanName -> 实现
}
// map 的 key 必须是 string，值为 beanName
```

## 说明

* slice 按注册顺序排列，map 以 beanName 为 key
* 不走 `BeanSelector`（全量收集，不选择）
* 支持 `aware:",omitempty"`：无候选时保持空 slice/map
