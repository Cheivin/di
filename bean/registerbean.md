---
description: 手动将一个实例化的对象注册进容器中
---

# 注册实例

### 注册普通实例

RegisterBean

| 参数名  | 类型          |      说明     |
| ---- | ----------- | :---------: |
| bean | interface{} | 将要注册bean的指针 |

```
dao:=&Dao{Prefix:"test"}
di.RegisterBean(dao)
```

### 注册命名实例

RegisterNamedBean

| 参数名      | 类型          |      说明     |
| -------- | ----------- | :---------: |
| beanName | string      | 指定的beanName |
| bean     | interface{} | 将要注册bean的指针 |

```
dao:=&Dao{Prefix:"test"}
di.RegisterNamedBean("dao",dao)
```

### 说明

* 注册实例需提供指针
* 普通实例的beanName由[beanname-generate-rule.md](../others/beanname-generate-rule.md "mention")确定。
* 命名实例的beanName由`RegisterNamedBean`的第一个参数确定。
* 可以动态注册bean，即在`Load()`方法执行后也可以使用。

