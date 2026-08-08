---
layout: default
title: 内置管理器 van
nav_order: 2
parent: 配置管理
---

# 内置管理器 van

`van` 是 di 内置的 `ValueStore` 实现，支持点号分隔的层级 key 和类型转换。

## 层级 key

key 用 `.` 分隔表示层级，`Set` map 类型的值会自动展开：

```go
vs := van.New()
vs.Set("db", map[string]any{
	"host": "localhost",
	"port": 3306,
	"pool": map[string]any{
		"max-idle": 10,
	},
})

vs.Get("db.host")         // "localhost"
vs.Get("db.port")         // 3306
vs.Get("db.pool.max-idle") // 10
```

## 大小写不敏感

所有 key 统一转小写存储，查询时不区分大小写：

```go
vs.Set("App.Port", 8080)
vs.Get("app.port") // 8080
vs.Get("APP.PORT") // 8080
```

## 合并语义

同名 key 的值合并：
- 两者都是 map：递归合并
- 类型冲突（如 string vs map）：**新值覆盖旧值**（v0.4.0 修正，原为静默丢弃）

```go
vs.Set("config", map[string]any{"a": 1, "b": 2})
vs.Set("config", map[string]any{"b": 3, "c": 4})
vs.Get("config") // map[a:1 b:3 c:4]（b 被新值覆盖，a 保留，c 新增）
```

## 类型转换（van.Cast）

`van.Cast(value, targetType)` 将值转为目标 reflect.Type：

```go
// 字符串转数值
van.Cast("42", reflect.TypeOf(int(0)))     // 42 (int)
van.Cast("3.14", reflect.TypeOf(float64(0))) // 3.14

// 字符串转 bool
van.Cast("true", reflect.TypeOf(false))    // true

// Duration：带单位直接解析，纯数字按毫秒兜底
van.Cast("5s", reflect.TypeOf(time.Second))    // 5s
van.Cast("5000", reflect.TypeOf(time.Second))  // 5000ms = 5s

// slice 转换（v0.4.0 新增）：元素逐个转换
van.Cast([]int{1,2,3}, reflect.TypeOf([]string{})) // ["1","2","3"]

// Stringer 接口（v0.4.0 新增）
type MyType struct{}
func (MyType) String() string { return "42" }
van.Cast(MyType{}, reflect.TypeOf(int(0))) // 42
```

## 未知类型兜底（v0.4.0）

v0.4.0 前，`toString` 对未知类型返回空字符串（静默丢值）。v0.4.0 起改用 `fmt.Sprint` 兜底，不再丢值。
