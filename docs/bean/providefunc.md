---
layout: default
title: 构造函数注入
nav_order: 3
parent: Bean 管理
---

# 构造函数注入

{: .new-title}
> v0.4.0 新增
>
> `ProvideFunc` 在 v0.4.0 引入，支持以工厂函数的方式注册 bean。

`ProvideFunc` 注册一个工厂函数，容器在 `Load()` 时按入参类型注入依赖，用函数的返回值作为 bean。它让 Go 也能写出类似 Spring "构造方法注入" 的代码。

## 何时使用

`Provide` + `aware` 标签适合字段依赖清晰、可无参构造的场景。当 bean 的创建**需要逻辑**时，用 `ProvideFunc` 更合适：

- 构造时需要根据依赖做计算（如拼连接串、初始化内部状态）
- 依赖需要校验、转换、包裹
- 不想把所有依赖都暴露为可导出字段

## API

```go
func ProvideFunc(fn any) DI
```

## 函数签名要求

`fn` 必须满足：

- 是 `func` 类型（非函数会 Fatal，错误包装 `ErrBean`）
- **返回值恰好 1 个**，且必须是**指针**（多返回值或返回值类型会 Fatal）

```go
// 正确：入参为依赖类型，单个指针返回值
func newUserDao(db *DB, cache *Cache) *UserDao {
    return &UserDao{DB: db, Cache: cache}
}
```

## 入参解析规则

容器对每个入参按以下顺序解析：

1. **指针入参**：先按类型推断的 beanName（`GetBeanName(*DB)` → `dB`）查找
2. **找不到 / 接口入参**：按类型匹配所有候选，交由 [BeanSelector](selector) 决定选中哪个（默认取最后注册的）

工厂入参也会参与[循环依赖检测](cycle-detection)。

## beanName 推断

工厂产物的 beanName 由返回值类型决定，规则同 `Provide`：优先 `BeanName()` 接口，否则取类型名首字母小写。

```go
func newUserDao(db *DB) *UserDao { ... }   // beanName → "userDao"
func newDao(db *DB) (d *Dao)     { ... }   // beanName → "dao"
```

## 生命周期

工厂产物与 `Provide` 实例化的 bean 一致，走完整生命周期：

```
工厂调用(返回指针) → BeanConstruct → PreInitialize → 注入(aware/value)
                 → AfterPropertiesSet → Initialized → Destroy(容器销毁时)
```

虽然工厂已经构造了对象，但 `aware` / `value` 字段仍会按标签注入，且 `BeanConstruct` 等回调照常触发。

## 示例

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct {
	Prefix string
}

type Cache struct{}

// UserDao 不使用 aware 标签，依赖通过工厂函数入参注入
type UserDao struct {
	DB    *DB
	Cache *Cache
}

// newUserDao 是工厂函数：入参类型即依赖类型，返回值作为 bean
func newUserDao(db *DB, cache *Cache) *UserDao {
	return &UserDao{DB: db, Cache: cache}
}

func main() {
	di.RegisterBean(&DB{Prefix: "tbl_"})
	di.RegisterBean(&Cache{})

	// ProvideFunc 注册工厂函数
	di.ProvideFunc(newUserDao)

	di.Load()

	dao, _ := di.GetBean("userDao")
	u := dao.(*UserDao)
	fmt.Printf("userDao.DB.Prefix = %q\n", u.DB.Prefix)   // "tbl_"
	fmt.Printf("userDao.Cache injected = %v\n", u.Cache != nil) // true
}
```

## 接口入参与多实现

工厂入参为接口类型时，容器按类型匹配所有实现，由 `BeanSelector` 选定一个：

```go
type Sender interface{ Send(string) }

// 入参是接口，容器会收集所有 Sender 实现并按策略选择
func newNotifier(sender Sender) *Notifier {
	return &Notifier{sender: sender}
}

func main() {
	c := di.New()
	c.WithBeanSelector(di.PrimaryFirst{}) // 用 PrimaryFirst 策略
	c.Provide(EmailSender{})              // 假设 EmailSender 是 Primary
	c.Provide(SmsSender{})
	c.ProvideFunc(newNotifier)
	c.Load()
}
```

详见 [接口选择策略](selector)。

## ProvideFunc vs Provide(aware)

| 维度 | `ProvideFunc` | `Provide` + `aware` |
|------|---------------|---------------------|
| 依赖来源 | 函数入参（构造期注入） | 字段标签（注入期赋值） |
| 适合场景 | 需要构造逻辑 / 依赖校验 | 字段依赖清晰、可无参构造 |
| 字段可见性 | 依赖可放私有字段（小写） | 必须是可导出字段（大写）才能注入 |
| 构造时机 | 工厂调用时一次性确定 | `Load()` 阶段逐字段赋值 |
| 生命周期 | 完整（BeanConstruct … Destroy） | 完整 |

## 错误处理

以下情况会 `Fatal`（panic，错误包装 `ErrBean`）：

- `fn` 不是函数
- 返回值数量不等于 1
- 返回值不是指针类型
- 入参依赖找不到对应 bean
- beanName 与已注册实例/定义冲突（`ErrBean` / `ErrDefinition`）

在 `Load()` 之后调用 `ProvideFunc` 会 Fatal（错误包装 `ErrLoaded`）。

## 相关

- [注册结构体](providebean) — 字段标签注入的对照方案
- [接口选择策略](selector) — 工厂接口入参如何选定实现
- [循环依赖检测](cycle-detection) — 工厂入参也参与检测
