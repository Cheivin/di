---
description: 将一个结构体注册进容器中，由容器托管其实例化
---

# 注册结构体

### 注册普通类

Provide

| 参数名       |      类型     | 说明               |
| --------- | :---------: | ---------------- |
| prototype | interface{} | 将要注册的结构体(推荐)或其指针 |

```
di.Provide(Dao{})
di.Provide(Service{})
```

### 注册命名类

ProvideNamedBean

| 参数名       | 类型          |        说明        |
| --------- | ----------- | :--------------: |
| beanName  | string      |    指定的beanName   |
| prototype | interface{} | 将要注册的结构体(推荐)或其指针 |

```
di.ProvideNamedBean("dao",Dao{})
di.ProvideNamedBean("service",Service{})
```

### 说明

* 普通类的beanName由[beanname-generate-rule.md](../others/beanname-generate-rule.md "mention")确定。
* 命名类的beanName由`ProvideNamedBean`的第一个参数确定。
* 只能在`Load()`方法执行前使用。
