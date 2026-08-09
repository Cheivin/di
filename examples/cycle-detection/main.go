// 示例：循环依赖
//
// 演示 di 的两阶段设计天然支持指针循环依赖，
// 以及如何通过 WithCircularCheck 开启严格检测。
package main

import (
	"errors"
	"fmt"

	"github.com/cheivin/di"
)

// A ↔ B 指针循环
type A struct {
	B *B `aware:""`
}

type B struct {
	A *A `aware:""`
}

func main() {
	fmt.Println("===== 默认行为：支持循环依赖 =====")
	c1 := di.New()
	c1.Provide(A{})
	c1.Provide(B{})
	c1.Load()

	a, _ := c1.GetBean("a")
	ra := a.(*A)
	fmt.Printf("A.B 注入成功=%v, A.B.A==A 循环闭环=%v\n", ra.B != nil, ra.B.A == ra)

	fmt.Println("\n===== opt-in：开启严格检测 =====")
	c2 := di.New()
	c2.WithCircularCheck(true) // 显式开启
	c2.Provide(A{})
	c2.Provide(B{})

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			if errors.Is(err, di.ErrCircularDependency) {
				fmt.Printf("检测到循环依赖: %v\n", err)
			}
		}
	}()
	c2.Load()
}
