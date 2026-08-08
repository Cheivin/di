---
layout: default
title: 错误处理
nav_order: 4
parent: 其他
---

# 错误处理

di 在遇到不可恢复的错误时（如依赖缺失、类型不匹配、循环依赖）会通过 logger 的 `Fatal` 触发 panic。

## Fatal 与 panic（v0.4.0 变更）

**v0.4.0 前**：`Fatal` 调用 `os.Exit(1)`，进程直接终止，无法捕获。

**v0.4.0 起**：`Fatal` 改为 panic，错误可被 `recover()` 捕获，且 panic 值是带 `Unwrap` 链的 `error`。

```go
type Log interface {
	// v0.4.0 前: Fatal(string)
	// v0.4.0 起: Fatal(error) — 接收 error，panic 它
	Fatal(error)
}
```

## errors.Is 判断错误类型

`Load()` 和 `Serve()` 会 panic，你可以 recover 后用 `errors.Is` 判断具体错误：

```go
defer func() {
	if r := recover(); r != nil {
		err := r.(error)
		switch {
		case errors.Is(err, di.ErrBean):
			// bean 相关错误（重复注册、依赖缺失、类型不匹配）
		case errors.Is(err, di.ErrDefinition):
			// bean 定义错误
		case errors.Is(err, di.ErrCircularDependency):
			// 循环依赖
		case errors.Is(err, di.ErrLoaded):
			// 容器已加载
		case errors.Is(err, di.ErrNotLoaded):
			// 容器未加载
		}
		fmt.Println(err)
	}
}()
di.Load()
```

## 错误哨兵值

| 变量 | 含义 |
|------|------|
| `ErrBean` | bean 相关错误（注册/注入/类型） |
| `ErrDefinition` | bean 定义错误（重复定义、notfound） |
| `ErrLoaded` | 容器已加载（重复 Load/Provide） |
| `ErrNotLoaded` | 容器未加载（未 Load 就 Serve） |
| `ErrCircularDependency` | 循环依赖（v0.4.0 新增） |

## 自定义 logger

如果不想用 panic 语义，可以实现自己的 `Log` 接口替换默认行为：

```go
type myLogger struct{}

func (myLogger) Fatal(err error) {
	// 改为记录日志后返回 error，或发送到监控系统
	log.Printf("di fatal: %v", err)
	panic(err) // 或自定义处理
}

c := di.New()
c.Log(myLogger{})
```

> 注意：`Fatal(error)` 签名是 v0.4.0 的 breaking change，自定义 logger 实现需同步修改。
