---
description: 一个接口多个实现时，通过 BeanSelector 策略和 Primary 标记控制选择（v0.4.0 新增）。
---

# 接口选择策略（BeanSelector）

当一个接口有多个实现、用 `aware:""` 单值注入时，由 `BeanSelector` 决定选谁。

## 配置策略

```go
c := di.New()
c.WithBeanSelector(di.PrimaryFirst{}) // 设置策略
```

## 内置策略

| 策略                   | 行为                            |
| -------------------- | ----------------------------- |
| `LastRegistered`（默认） | 取最后注册的，兼容历史行为                 |
| `FirstRegistered`    | 取第一个注册的                       |
| `PrimaryFirst`       | 优先取实现了 `Primary` 接口的；无则取最后注册的 |
| `ErrorOnAmbiguous`   | 候选超过 1 个直接报错（严格模式）            |

## Primary 标记

bean 实现 `IsPrimary() bool` 声明自己是首选：

```go
type SmsSender struct{}
func (SmsSender) IsPrimary() bool { return true } // 首选

type EmailSender struct{}

type Notifier struct {
	Sender Sender `aware:""` // PrimaryFirst 时选 SmsSender
}
```

## 自定义策略

实现 `BeanSelector` 接口：

```go
type MySelector struct{}
func (MySelector) Select(candidates []di.BeanWithName, target reflect.Type) (int, error) {
	// 自定义选择逻辑，返回索引；error 非 nil 触发 Fatal
	return 0, nil
}
```
