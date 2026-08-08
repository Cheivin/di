---
description: Load 时自动检测循环依赖并给出清晰的错误链路（v0.4.0 新增）。
---

# 循环依赖检测

`Load()` 时自动做拓扑检测，发现 aware 依赖或 `ProvideFunc` 工厂入参构成环时直接 panic。

## 示例

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
c.Load() // panic: circular dependency: a -> b -> a
```

## 错误处理

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

## 检测范围

* `aware` 标签声明的依赖（命名/接口）
* `ProvideFunc` 工厂函数的入参依赖
* `value` 配置项不参与（不构成 bean 依赖）
