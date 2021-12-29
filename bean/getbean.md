---
description: 从容器中获取被托管的实例，不论是手动注册的实例还是结构体
---

# 获取bean

### 根据BeanName获取

GetBean

| 参数名      | 类型     |  说明 |
| -------- | ------ | :-: |
| beanName | string |  名称 |

| 序号 | 类型          |    说明    |
| -- | ----------- | :------: |
| 0  | interface{} | 获取到的实例指针 |
| 1  | bool        |  是否成功获取  |

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

| 序号 | 类型          |    说明    |
| -- | ----------- | :------: |
| 0  | interface{} | 获取到的实例指针 |
| 1  | bool        |  是否成功获取  |

```
bean, ok := di.GetByType(Service{})
if !ok {
	panic("service 不存在")
}
service := bean.(*Service)
```

