---
layout: default
title: 概述
nav_order: 1
parent: 入门
---

# 概述

`di` 是一个简易版本的 Go 依赖注入（Dependency Injection）实现，灵感来自 Java 的 Spring IoC 容器。

## 它解决什么问题

在没有依赖注入框架时，组件之间的依赖关系需要手动维护：

```go
// 传统写法：手动组装依赖
db := NewDB()
repo := NewUserRepo(db)
service := NewUserService(repo)
handler := NewHandler(service)
```

当依赖关系复杂时，这种手动组装变得难以维护。`di` 让你声明依赖（通过标签），由容器自动组装：

```go
type UserService struct {
	Repo *UserRepo `aware:""` // 声明依赖，容器自动注入
}
```

## 核心概念

| 概念 | 说明 |
|------|------|
| **Bean** | 由容器托管的对象实例 |
| **注册** | `RegisterBean`（已有实例）或 `Provide`（类型原型） |
| **注入** | 通过 `aware` 标签声明依赖，容器自动赋值 |
| **生命周期** | BeanConstruct → PreInitialize → 注入 → AfterPropertiesSet → Initialized → Destroy |
| **配置** | 通过 `value` 标签注入配置项，支持类型转换 |

## 特性一览

- 手动注册 bean 实例 / 注册结构体原型自动实例化
- 构造函数注入（`ProvideFunc`）
- slice / map 批量注入
- 接口歧义策略（`BeanSelector` / `Primary`）
- 配置项注入与类型自动转换
- 完整生命周期管理
- 循环依赖检测
- 线程安全，支持运行期动态注册
