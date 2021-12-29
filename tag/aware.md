---
description: 用于标记需要依赖注入的字段
---

# aware

### 使用

修饰的字段必须为类型指针。

#### 普通注入

* Tag格式为 `aware:""`，引号中内容为空字符串。
* 此时会根据[beanname-generate-rule.md](../others/beanname-generate-rule.md "mention")得到目标字段的beanName注入。
* 如果依赖的bean未找到，会引发panic

#### 按名称注入

* Tag格式为 `aware:"beanName"`，引号中填写依赖的bean名称。
* 如果依赖的bean未找到，会引发panic

#### 修饰非必须依赖项

* Tag格式为 `aware:"omitempty"` 或 `aware:"beanName,omitempty"`。分别对应普通注入和按名称注入两种情况。
* 如果依赖的bean未找到，不会引发panic，而是填充nil值

