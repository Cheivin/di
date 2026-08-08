// 示例：接口歧义策略（BeanSelector / Primary）
//
// 演示一个接口有多个实现时，如何控制单值注入选中哪个。
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
	n1.(*Notifier).Sender.Send("hello") // 默认取 smsSender（最后注册）

	fmt.Println("\n===== PrimaryFirst 策略（优先 Primary 实现）=====")
	c2 := di.New()
	c2.WithBeanSelector(di.PrimaryFirst{})
	c2.Provide(EmailSender{}) // Primary
	c2.Provide(SmsSender{})
	c2.Provide(Notifier{})
	c2.Load()

	n2, _ := c2.GetBean("notifier")
	n2.(*Notifier).Sender.Send("hello") // PrimaryFirst 取 emailSender（它是 Primary）
}
