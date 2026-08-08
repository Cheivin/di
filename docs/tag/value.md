---
layout: default
title: value 标签
nav_order: 2
parent: 标签
---

# value 标签

`value` 标签声明结构体字段从配置项注入值，支持类型自动转换。

## 基本用法

```go
type AppConfig struct {
	Port    int    `value:"app.port"`
	Host    string `value:"app.host"`
	Debug   bool   `value:"app.debug"`
	Timeout time.Duration `value:"app.timeout"`
}
```

配置通过 `SetProperty` / `SetDefaultProperty` 设置：

```go
di.SetDefaultProperty("app.port", 8080)
di.SetProperty("app.host", "0.0.0.0")
di.SetProperty("app.debug", "true")    // 字符串转 bool
di.SetProperty("app.timeout", "30s")   // Duration 带单位
di.Provide(AppConfig{})
di.Load()
```

## 支持的类型

| 目标类型 | 转换方式 |
|---------|---------|
| `string` | 任意值转字符串（支持 `fmt.Stringer`） |
| `bool` | `strconv.ParseBool`（"true"/"false"/"1"/"0"） |
| `int`/`int8`...`int64` | `strconv.ParseInt` |
| `uint`/`uint8`...`uint64` | `strconv.ParseUint` |
| `float32`/`float64` | `strconv.ParseFloat` |
| `time.Duration` | `time.ParseDuration`；**纯数字按毫秒兜底** |
| `[]T`（v0.4.0） | slice 元素逐个转换 |

## Duration 的特殊处理

`time.Duration` 本质是 `int64`，配置注入有两种形式：

```go
di.SetProperty("timeout", "30s")    // 带单位，直接 ParseDuration
di.SetProperty("retry", "5000")     // 纯数字 → 5000ms（毫秒兜底，历史兼容行为）
```

## 缺省值

配置项不存在时，字段保持零值，**不会报错**：

```go
type Config struct {
	Optional string `value:"missing.key"` // missing.key 未设置，Optional = ""
}
```

## 配置层级 key

配置项支持点号分隔的层级（见 [van 配置管理器](../valuestore/van)）：

```go
di.SetDefaultProperty("db", map[string]any{
	"host": "localhost",
	"port": 3306,
	"pool": map[string]any{
		"max-idle": 10,
	},
})

type DBConfig struct {
	Host    string `value:"host"`
	Port    int    `value:"port"`
	MaxIdle int    `value:"pool.max-idle"`
}

// LoadProperties 加载到独立结构体（不注册为 bean）
cfg := di.LoadProperties("db.", &DBConfig{}).(DBConfig)
// cfg.Host = "localhost", cfg.Port = 3306, cfg.MaxIdle = 10
```

## 与 aware 的区别

| 标签 | 注入内容 | 来源 |
|------|---------|------|
| `aware` | bean 实例 | 容器注册的 bean |
| `value` | 配置值（基础类型） | ValueStore 配置存储 |

一个字段只能用一个标签。
