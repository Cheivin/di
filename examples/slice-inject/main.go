// 示例：slice / map 批量注入
//
// 演示 []T 和 map[string]T 字段加 aware:"" 标签，自动收集所有实现。
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

// Handler 是一个接口，有多个实现
type Handler interface {
	Handle() string
}

type LogHandler struct{}

func (LogHandler) Handle() string   { return "log" }
func (LogHandler) BeanName() string { return "logHandler" }

type AuthHandler struct{}

func (AuthHandler) Handle() string   { return "auth" }
func (AuthHandler) BeanName() string { return "authHandler" }

// Router 通过 slice 收集所有 Handler，通过 map 以 beanName 为 key 收集
type Router struct {
	Handlers []Handler          `aware:""` // 按注册顺序
	ByName   map[string]Handler `aware:""` // beanName -> 实现
}

func main() {
	di.Provide(LogHandler{})
	di.Provide(AuthHandler{})
	di.Provide(Router{})

	di.Load()

	router, _ := di.GetBean("router")
	r := router.(*Router)

	fmt.Printf("收集到 %d 个 handler（按注册顺序）:\n", len(r.Handlers))
	for _, h := range r.Handlers {
		fmt.Printf("  - %s\n", h.Handle())
	}

	fmt.Printf("\nmap 收集 %d 个（按 beanName）:\n", len(r.ByName))
	for name, h := range r.ByName {
		fmt.Printf("  %s -> %s\n", name, h.Handle())
	}
}
