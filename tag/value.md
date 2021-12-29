---
description: 用来标记需要注入的配置项
---

# value

### 使用

* 该Tag的完整格式为 `value:"property"`，property为配置项名称
* 该Tag可以标记的字段类型：
  * 基本数据类型
    * string
    * bool
    * int、int64、int32、int16、int8
    * uint、uint64、uint32、uint16、uint8
    * float64、float32
  * 其他类型
    * time.Duration

### 说明

* 注入的配置由`配置项管理器`管理，相关方法参见 [definition.md](../valuestore/definition.md "mention")。
* `time.Duration` 类型默认单位为`ms`，即 `100` 注入后对应类型值应为 `100ms`。

