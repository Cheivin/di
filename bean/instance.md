---
description: 从容器中创建一个新的实例，仅支持注册的结构体
---

# 手动实例化

### 根据BeanName创建

NewBeanByName

| 参数名      | 类型     |  说明 |
| -------- | ------ | :-: |
| beanName | string |  名称 |

<table><thead><tr><th width="150">序号</th><th align="center">类型</th><th>说明</th></tr></thead><tbody><tr><td>0</td><td align="center">interface{}</td><td>创建的新实例指针</td></tr></tbody></table>

```
bean := di.NewBeanByName("service")
service := bean.(*Service)
```

### 根据类型创建

NewBean

| 参数名      | 类型          |  说明  |
| -------- | ----------- | :--: |
| beanType | interface{} | 实例类型 |

<table><thead><tr><th width="150">序号</th><th align="center">类型</th><th>说明</th></tr></thead><tbody><tr><td>0</td><td align="center">interface{}</td><td>获取到的实例指针</td></tr></tbody></table>

```
bean := di.NewBean(Service{})
service := bean.(*Service)
```

### 说明

* 根据名称创建时，如果beanName对应的结构体不存在，会引发panic。
* 根据类型创建时，如果类型未受容器托管，依然会根据bean的生命周期创建，但类型不会注册进容器中。
* 创建的实例不会受容器托管。
