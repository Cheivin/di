---
layout: default
title: 接口选择策略
nav_order: 6
parent: Bean 管理
---

# 接口选择策略

{: .new-title}
> v0.4.0 新增
>
> `BeanSelector` 接口与 `Primary` 机制在 v0.4.0 引入。

## 问题：一个接口多个实现

当一个接口字段用单值 `aware:""` 注入，而容器里注册了多个实现时，该选哪个？

```go
type Sender interface{ Send(string) }

type EmailSender struct{}
type SmsSender struct{}

type Notifier struct {
	Sender Sender `aware:""` // EmailSender 和 SmsSender 都实现 Sender，选谁？
}
```

容器内没有名称线索时（`aware:""` 空名称按类型推断），会按类型匹配出所有候选，再交由 **`BeanSelector`** 决定。

## BeanSelector 接口

```go
type BeanSelector interface {
    Select(candidates []BeanWithName, targetType reflect.Type) (int, error)
}
```

- `candidates`：按 `beanSort` 注册顺序排列，至少 1 个元素
- 返回选中索引；返回 error 会触发 `Fatal`（panic，错误包装 `ErrBean`）

通过 `WithBeanSelector` 配置全局策略：

```go
c := di.New()
c.WithBeanSelector(di.PrimaryFirst{}) // 设置策略
```

## 内置策略

| 策略 | 行为 | 适用场景 |
|------|------|---------|
| `LastRegistered` | 取最后注册的候选（**默认**） | 兼容历史行为，后注册覆盖前者 |
| `FirstRegistered` | 取第一个注册的候选 | 优先级以注册顺序为准 |
| `PrimaryFirst` | 优先 `Primary` 实现；无则回退到最后注册 | 想显式声明首选实现 |
| `ErrorOnAmbiguous` | 候选超过 1 个直接报错 | 严格模式，强制显式命名注入 |

### LastRegistered（默认）

取候选列表的最后一个。这是默认值，与 v0.4.0 之前的行为完全一致——后注册的 bean 覆盖前者。

```go
c.Provide(EmailSender{}) // 先注册
c.Provide(SmsSender{})   // 后注册
// Notifier.Sender → SmsSender
```

### FirstRegistered

取第一个。与 `LastRegistered` 相反，先注册的优先。

### PrimaryFirst

优先选中实现了 `Primary` 接口（`IsPrimary() bool` 返回 true）的候选：

- 恰好一个 Primary：返回它
- 多个 Primary：返回 error（歧义）
- 无 Primary：回退到最后注册的（兼容默认）

### ErrorOnAmbiguous

严格模式：候选超过 1 个直接报错，要求显式命名注入或用 `Primary`。适用于不希望隐式依赖"最后注册"语义的项目。

## Primary 接口

bean 实现 `Primary` 接口，声明自己为首选实现：

```go
type Primary interface {
    IsPrimary() bool
}
```

只有 `PrimaryFirst` 策略会识别它；其他策略忽略。`IsPrimary()` 返回 false 等同于未实现。

## 示例：默认 vs PrimaryFirst

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type Sender interface {
	Send(msg string)
}

type EmailSender struct{}

func (EmailSender) Send(msg string)  { fmt.Printf("  [email] %s\n", msg) }
func (EmailSender) BeanName() string { return "emailSender" }
func (EmailSender) IsPrimary() bool  { return true } // Email 是首选

type SmsSender struct{}

func (SmsSender) Send(msg string)  { fmt.Printf("  [sms] %s\n", msg) }
func (SmsSender) BeanName() string { return "smsSender" }

// Notifier 单值注入 Sender，多实现时由策略决定
type Notifier struct {
	Sender Sender `aware:""`
}

func main() {
	fmt.Println("===== 默认策略 LastRegistered（取最后注册的）=====")
	c1 := di.New()
	c1.Provide(EmailSender{})
	c1.Provide(SmsSender{})
	c1.Provide(Notifier{})
	c1.Load()

	n1, _ := c1.GetBean("notifier")
	n1.(*Notifier).Sender.Send("hello") // 输出: [sms] hello（默认取最后注册的）

	fmt.Println("\n===== PrimaryFirst 策略（优先 Primary 实现）=====")
	c2 := di.New()
	c2.WithBeanSelector(di.PrimaryFirst{})
	c2.Provide(EmailSender{}) // Primary
	c2.Provide(SmsSender{})
	c2.Provide(Notifier{})
	c2.Load()

	n2, _ := c2.GetBean("notifier")
	n2.(*Notifier).Sender.Send("hello") // 输出: [email] hello（PrimaryFirst 选 Primary）
}
```

## 自定义策略

实现 `BeanSelector` 接口即可。例如按 beanName 字典序选择：

```go
type ByName struct{}

func (ByName) Select(candidates []di.BeanWithName, _ reflect.Type) (int, error) {
	best := 0
	for i := 1; i < len(candidates); i++ {
		if candidates[i].Name < candidates[best].Name {
			best = i
		}
	}
	return best, nil
}

c.WithBeanSelector(ByName{})
```

## 替代方案：显式命名注入

如果不想依赖选择策略，可以直接在 `aware` 标签里指定 beanName，绕开歧义：

```go
type Notifier struct {
	Sender Sender `aware:"emailSender"` // 显式指定，不走策略
}
```

命名注入只查找精确匹配的 bean，找不到才报错，不会触发 `BeanSelector`。

## 相关

- [批量注入](slice-inject) — 需要全部实现而非选一个时
- [构造函数注入](providefunc) — 工厂入参为接口时同样走选择策略
- [获取 bean](getbean) — `GetByType` 单值获取的歧义处理
