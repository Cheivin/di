// 示例：循环依赖检测
//
// 演示 Load() 时自动检测循环依赖并给出清晰的错误链路。
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

// A → B → A 构成循环
type A struct {
	B *B `aware:""`
}

type B struct {
	A *A `aware:""`
}

// 正常的依赖链 C → D（无环）
type C struct {
	D *D `aware:""`
}

type D struct{}

func main() {
	fmt.Println("===== 正常依赖（无环）=====")
	c1 := di.New()
	c1.Provide(C{})
	c1.Provide(D{})
	c1.Load()
	fmt.Println("加载成功")

	fmt.Println("\n===== 循环依赖（A → B → A）=====")
	c2 := di.New()
	c2.Provide(A{})
	c2.Provide(B{})

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("检测到循环依赖: %v\n", r)
			// 可用 errors.Is 判断
			// err := r.(error)
			// errors.Is(err, di.ErrCircularDependency)
		}
	}()
	c2.Load()
}
