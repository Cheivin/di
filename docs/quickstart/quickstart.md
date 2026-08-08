---
layout: default
title: 快速入门
nav_order: 2
parent: 入门
---

# 快速入门

## 1. 声明你的类型

```go
type DB struct {
	Prefix string
}

// UserService 声明对 DB 的依赖
type UserService struct {
	DB *DB `aware:"db"` // aware 标签指定依赖的 bean 名称
}
```

## 2. 注册并加载

```go
func main() {
	// 注册一个已实例化的 bean，命名为 "db"
	di.RegisterNamedBean("db", &DB{Prefix: "tbl_"})

	// 注册结构体原型，由容器实例化
	di.Provide(UserService{})

	// 加载容器（触发实例化、依赖注入、生命周期回调）
	di.Load()

	// 获取 bean
	svc, _ := di.GetBean("userService")
	fmt.Println(svc.(*UserService).DB.Prefix) // "tbl_"
}
```

## 3. 更简洁的写法：按类型推断

如果依赖字段的类型在容器中只有一个实现，可以省略 bean 名称：

```go
type UserService struct {
	DB *DB `aware:""` // 空名，按类型推断 beanName = "dB"
}
```

## 4. 全局函数 vs 容器实例

上面的示例用了全局函数（`di.RegisterBean` 等），它们操作一个全局容器。你也可以创建独立容器：

```go
c := di.New()
c.RegisterNamedBean("db", &DB{Prefix: "tbl_"})
c.Provide(UserService{})
c.Load()

svc, _ := c.GetBean("userService")
```

> 全局容器是**懒初始化**的：首次调用全局函数时才创建，`di.Reset()` 可重置（主要用于测试隔离）。

## 下一步

- [Bean 注册](../bean/registerbean) — 了解所有注册方式
- [aware 标签](../tag/aware) — 依赖注入详解
- [构造函数注入](../bean/providefunc) — v0.4.0 新增的工厂注入
- [示例代码](https://github.com/cheivin/di/tree/main/examples) — 完整可运行示例
