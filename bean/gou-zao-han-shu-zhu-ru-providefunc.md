---
description: 通过工厂函数注册 bean，容器按入参类型注入依赖（v0.4.0 新增）。
---

# 构造函数注入（ProvideFunc）

`ProvideFunc` 注册一个工厂函数，容器按入参类型自动注入依赖，用返回值作为 bean。

## 基本用法

```go
type UserDao struct {
	DB    *DB
	Cache *Cache
}

// 工厂函数：入参即依赖，返回值作为 bean
func newUserDao(db *DB, cache *Cache) *UserDao {
	return &UserDao{DB: db, Cache: cache}
}

func main() {
	di.RegisterBean(&DB{Prefix: "tbl_"})
	di.RegisterBean(&Cache{})
	di.ProvideFunc(newUserDao) // 注册工厂
	di.Load()
}
```

## 函数签名要求

* 必须是 `func(...) (...)`
* 返回值**恰好 1 个**，且为**指针类型**
* beanName：优先用返回类型实现 `BeanName()` 接口的返回值，否则按类型推断（首字母小写）

## 入参注入规则

1. 指针入参：按类型推断 beanName 查找
2. 接口入参：按类型匹配所有实现，多个时由 `BeanSelector` 策略决定

## 与 Provide(aware 标签) 的对比

| 方式                          | 适合场景                |
| --------------------------- | ------------------- |
| `Provide(T{})` + `aware` 标签 | 简单结构体，依赖用标签声明       |
| `ProvideFunc(fn)`           | 需要构造逻辑（参数计算、条件初始化等） |

工厂产物仍会走完整生命周期（`BeanConstruct`/`AfterPropertiesSet`/`Initialized`/`Destroy`）。
