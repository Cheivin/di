---
layout: default
title: 配置管理器接口
nav_order: 1
parent: 配置管理
---

# 配置管理器接口

配置存储由 `ValueStore` 接口抽象，di 默认使用内置的 `van` 实现。

## ValueStore 接口

```go
type ValueStore interface {
	SetDefault(key string, value any) // 设置默认值（低优先级）
	Set(key string, value any)        // 设置值（高优先级，覆盖 SetDefault）
	Get(key string) any               // 获取值（优先 Set，其次 SetDefault）
	GetAll() map[string]any           // 获取所有配置
}
```

## 优先级

`Set` 的值覆盖 `SetDefault`：

```go
vs := van.New()
vs.SetDefault("port", 8080)
vs.Set("port", 9090)
vs.Get("port") // 9090（Set 覆盖 SetDefault）
```

## 替换实现

你可以实现自己的 `ValueStore`（比如从远程配置中心加载）：

```go
type RemoteConfigStore struct { /* ... */ }

func (r *RemoteConfigStore) SetDefault(key string, value any) { /* ... */ }
func (r *RemoteConfigStore) Set(key string, value any)        { /* ... */ }
func (r *RemoteConfigStore) Get(key string) any               { /* 从远程获取 */ }
func (r *RemoteConfigStore) GetAll() map[string]any           { /* ... */ }

c := di.New()
c.UseValueStore(&RemoteConfigStore{})
```

## 容器上的配置方法

`DI` 接口提供了便捷方法，转发到底层 `ValueStore`：

| 方法 | 等价于 |
|------|--------|
| `SetDefaultProperty(k, v)` | `valueStore.SetDefault(k, v)` |
| `SetProperty(k, v)` | `valueStore.Set(k, v)` |
| `GetProperty(k)` | `valueStore.Get(k)` |
| `SetDefaultPropertyMap(m)` | 遍历 m 调 SetDefault |
| `SetPropertyMap(m)` | 遍历 m 调 Set |
| `LoadProperties(prefix, type)` | 按 prefix 加载配置到结构体 |

详见 [van 配置管理器](./van)。
