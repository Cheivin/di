# di

`di` 是一个简易版本的 Go 依赖注入实现，类似 Spring IoC 容器。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 特性

- ✅ 手动注册 bean 实例 / 注册结构体原型自动实例化
- ✅ **构造函数注入**（`ProvideFunc`）：工厂函数按入参类型注入依赖
- ✅ 按名称 / 类型获取 bean，按类型手动实例化
- ✅ **slice / map 批量注入**：收集同接口的所有实现
- ✅ **接口歧义策略**：`BeanSelector` / `Primary` 多实现选择
- ✅ **只读管理/诊断 API**：`GetBeanNames` / `DescribeBean` / `GetBeanDependencies`（v0.6.2 新增）
- ✅ 配置项注入（`value` 标签），支持类型自动转换
- ✅ 匿名字段注入
- ✅ **完整生命周期管理**（构造 → 注入前 → 注入后 → 初始化 → 销毁）
- ✅ **循环依赖检测**，启动期快速定位问题
- ✅ **线程安全**，支持运行期动态注册与并发读取

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct{}

type UserService struct {
	DB *DB `aware:"db"`
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(UserService{})
	di.Load()

	svc, _ := di.GetBean("userService")
	fmt.Println(svc.(*UserService).DB != nil) // true
}
```

## 文档目录

### 入门

- [概述](quickstart/introduction) — di 是什么、核心概念
- [快速入门](quickstart/quickstart) — 5 分钟上手
- [安装](quickstart/install) — 安装与版本要求

### Bean 管理

- [注册实例](bean/registerbean) — RegisterBean 手动注册
- [注册结构体](bean/providebean) — Provide 原型注册
- [构造函数注入](bean/providefunc) — ProvideFunc 工厂注入（v0.4.0 新增）
- [获取 bean](bean/getbean) — GetBean / GetByType / 管理诊断 API（v0.6.2 新增）
- [批量注入](bean/slice-inject) — slice/map 收集同类型 bean（v0.4.0 新增）
- [接口选择策略](bean/selector) — BeanSelector / Primary（v0.4.0 新增）
- [生命周期](bean/lifecycle) — 完整生命周期回调
- [循环依赖检测](bean/cycle-detection) — 启动期自动检测（v0.4.0 新增）

### 标签

- [aware](tag/aware) — bean 依赖注入标签
- [value](tag/value) — 配置项注入标签

### 配置管理

- [配置管理器接口](valuestore/definition) — ValueStore 接口
- [内置管理器 van](valuestore/van) — van 实现与类型转换

### 其他

- [UnsafeMode](others/unsafemode) — 不安全模式（私有字段注入）
- [beanName 生成策略](others/beanname) — 名称推断规则
- [注入匿名字段](others/anonymous) — 匿名字段注入
- [错误处理](others/error-handling) — Fatal 与 errors.Is（v0.4.0 变更）
- [并发安全](others/concurrency) — 线程安全保证（v0.4.0 新增）
- [全局容器与 Reset](others/global) — 懒初始化与测试隔离

### 参考

- [更新日志](https://github.com/cheivin/di/blob/main/CHANGELOG.md)
- [示例代码](https://github.com/cheivin/di/tree/main/examples)
