---
description: Fatal 错误处理机制与 errors.Is 判断（v0.4.0 变更）。
---

# 错误处理

di 在遇到不可恢复的错误时通过 logger 的 `Fatal` 触发 panic。

## v0.4.0 变更

**v0.4.0 前**：`Fatal` 调用 `os.Exit(1)`，进程直接终止，无法捕获。

**v0.4.0 起**：`Fatal` 改为 panic，错误可被 `recover()` 捕获，panic 值是带 `Unwrap` 链的 `error`。

## Log 接口变更

```go
type Log interface {
	// v0.4.0 前: Fatal(string)
	// v0.4.0 起: Fatal(error)
	Fatal(error)
}
```

## errors.Is 判断错误类型

```go
defer func() {
	if r := recover(); r != nil {
		err := r.(error)
		switch {
		case errors.Is(err, di.ErrBean):               // bean 相关错误
		case errors.Is(err, di.ErrDefinition):         // bean 定义错误
		case errors.Is(err, di.ErrCircularDependency): // 循环依赖
		case errors.Is(err, di.ErrLoaded):             // 容器已加载
		case errors.Is(err, di.ErrNotLoaded):          // 容器未加载
		}
	}
}()
di.Load()
```

## 错误哨兵值

| 变量                      | 含义                  |
| ----------------------- | ------------------- |
| `ErrBean`               | bean 相关错误（注册/注入/类型） |
| `ErrDefinition`         | bean 定义错误           |
| `ErrLoaded`             | 容器已加载               |
| `ErrNotLoaded`          | 容器未加载（v0.4.0 新增）    |
| `ErrCircularDependency` | 循环依赖（v0.4.0 新增）     |
