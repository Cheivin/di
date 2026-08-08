---
layout: default
title: 注入匿名字段
nav_order: 3
parent: 其他
---

# 注入匿名字段

匿名字段（嵌入式结构体）也可以通过 `aware` 标签注入。

## 基本用法

```go
type Base struct {
	CreatedAt time.Time
}

type Service struct {
	*Base `aware:""` // 匿名嵌入，注入 Base 实例
	DB    *DB `aware:"db"`
}
```

## 限制

匿名字段注入的 bean **不能实现生命周期接口**（BeanConstruct/PreInitialize/AfterPropertiesSet/Initialized/Disposable），否则方法会被 Go 的方法提升规则意外触发，导致歧义。

容器在注册时会检测并报错：

```
fatal: error bean: base(*main.Base) as anonymous field in service(main.Service.base)
can not implements BeanConstruct
```

如果 Base 必须有生命周期逻辑，改为**非匿名字段**（命名字段）：

```go
type Service struct {
	Base *Base `aware:"base"` // 命名字段，允许生命周期
}
```
