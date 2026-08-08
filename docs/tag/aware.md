---
layout: default
title: aware 标签
nav_order: 1
parent: 标签
---

# aware 标签

`aware` 标签声明结构体字段需要容器注入依赖 bean。

## 基本用法

```go
type UserService struct {
	DB *DB `aware:"db"` // 注入名为 "db" 的 bean
}
```

## beanName 推断规则

| 标签写法 | 推断方式 |
|---------|---------|
| `aware:"db"` | 直接用 "db" 作为 beanName |
| `aware:""` | 空名，按类型推断：优先用字段类型实现 `BeanName()` 接口的返回值，否则用 `GetBeanName(类型)`（首字母小写） |

```go
type UserService struct {
	DB *DB `aware:""` // beanName = "dB"（DB 首字母小写）
}
```

## 支持的字段类型

### 指针字段（`*T`）

注入指定名称或同类型的指针 bean：

```go
type Service struct {
	DB *DB `aware:"db"`
}
```

> 指针字段不能指向接口（`*Interface` 会被拒绝）。

### 接口字段

注入实现该接口的 bean。空名时按类型匹配所有实现，多个实现时由 [BeanSelector 策略](../bean/selector)决定：

```go
type Service struct {
	Repo DataRepo `aware:""` // DataRepo 是接口，注入其实现
}
```

### slice / map 字段（v0.4.0 新增）

收集所有可赋值给元素类型的 bean，详见 [批量注入](../bean/slice-inject)：

```go
type Router struct {
	Handlers []Handler          `aware:""` // 收集所有 Handler 实现
	ByName   map[string]Handler `aware:""` // beanName -> 实现
}
```

## 可选注入：omitempty

`aware:"name,omitempty"` 或 `aware:",omitempty"` 表示依赖缺失时不报错（字段保持零值）：

```go
type Service struct {
	Cache *Cache `aware:",omitempty"` // Cache 不存在也不报错
}
```

## 匿名字段

匿名字段（嵌入式）也可以注入：

```go
type Service struct {
	*BaseService `aware:""` // 匿名嵌入
}
```

> 匿名字段的 bean 不能实现生命周期接口（BeanConstruct 等），否则方法会被意外提升导致歧义。容器会在注册时报错。

## 注入时机

注入发生在 `Load()` 阶段，顺序为：
1. 实例化所有 bean（`reflect.New`）
2. 触发 `BeanConstruct`（此时依赖尚未注入）
3. 触发 `PreInitialize`
4. **注入 aware 依赖**（本标签生效）
5. 触发 `AfterPropertiesSet`

因此 `BeanConstruct` 回调里访问 aware 字段会是 nil，依赖注入完成后才能使用。
