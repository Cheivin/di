---
description: 从容器中获取被托管的实例，不论是手动注册的实例还是结构体
---

# 获取bean

### 根据BeanName获取

GetBean

| 参数名      | 类型     |  说明 |
| -------- | ------ | :-: |
| beanName | string |  名称 |

<table><thead><tr><th width="150">序号</th><th align="center">类型</th><th>说明</th></tr></thead><tbody><tr><td>0</td><td align="center">interface{}</td><td>获取到的实例指针</td></tr><tr><td>1</td><td align="center">bool</td><td>是否成功获取</td></tr></tbody></table>

```
bean, ok := di.GetBean("service")
if !ok {
	panic("service 不存在")
}
service := bean.(*Service)
```

### 根据类型获取

GetByType

| 参数名      | 类型          |  说明  |
| -------- | ----------- | :--: |
| beanType | interface{} | 实例类型 |

<table><thead><tr><th width="150">序号</th><th align="center">类型</th><th>说明</th></tr></thead><tbody><tr><td>0</td><td align="center">interface{}</td><td>获取到的实例指针</td></tr><tr><td>1</td><td align="center">bool</td><td>是否成功获取</td></tr></tbody></table>

```
bean, ok := di.GetByType(Service{})
if !ok {
	panic("service 不存在")
}
service := bean.(*Service)
```

