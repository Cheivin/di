---
layout: default
title: beanName 生成策略
nav_order: 2
parent: 其他
---

# beanName 生成策略

bean 的名称（beanName）决定了它在容器中的唯一标识，也是 `aware` 标签匹配依赖的依据。

## 生成优先级

```
1. RegisterNamedBean / ProvideNamedBean 显式指定名称
2. 实现 BeanName 接口 → 用接口返回值
3. 默认：类型名首字母小写
```

## BeanName 接口

```go
type BeanName interface {
	BeanName() string
}
```

实现该接口的类型，注册时用接口返回值作为 beanName：

```go
type DB struct{}

func (*DB) BeanName() string {
	return "db" // 注册后 beanName = "db"，而非 "dB"
}

di.RegisterBean(&DB{})     // beanName = "db"
di.GetBean("db")           // ✅
di.GetBean("dB")           // ❌ not found
```

## 默认规则：首字母小写

未实现 `BeanName` 且未显式命名时，取类型名首字母小写：

| 类型 | beanName |
|------|----------|
| `UserService` | `userService` |
| `DB` | `dB` |
| `GormConfig` | `gormConfig` |

> 注意 `DB` → `dB`、`ID` → `iD`，只小写第一个字母。如需自定义，实现 `BeanName` 接口。

## aware 标签匹配

```go
type Service struct {
	A *DB `aware:"db"`  // 找 beanName == "db"
	B *DB `aware:""`    // 按类型推断：BeanName 接口 → "db"，否则 "dB"
}
```
