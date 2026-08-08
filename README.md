---
description: di 是一个简易版本的 Go 依赖注入实现，支持 bean 注册、自动依赖注入、配置注入、生命周期管理和条件装配。
---

# 概述

`di` 是一个简易版本的 Go 依赖注入（Dependency Injection）实现，类似 Spring IoC 容器。

### 特性

* 支持手动注册 bean 实例（`RegisterBean`）
* 支持注册 bean 类型原型，由 DI 容器自动实例化并托管（`Provide`）
* 支持构造函数注入，按工厂入参类型自动注入依赖（`ProvideFunc`）\[v0.4.0]
* 支持根据名称、类型获取 DI 容器托管的 bean 实例
* 支持根据类型手动生成新的 bean 实例并返回
* 支持配置项注入并转换成对应的基本类型
* 支持匿名字段的 bean 注入
* 支持 slice / map 批量注入，收集同类型的所有 bean \[v0.4.0]
* 支持接口多实现的选择策略（`BeanSelector` / `Primary`）\[v0.4.0]
* 支持完整生命周期管理（构造 / 注入前 / 注入后 / 初始化 / 销毁）
* 支持循环依赖检测，启动期快速定位问题 \[v0.4.0]
* 线程安全，支持运行期动态注册与并发读取 \[v0.4.0]

### 版本要求

Go 1.25+（v0.4.0 起）

### 安装

```bash
go get github.com/cheivin/di@latest
```
