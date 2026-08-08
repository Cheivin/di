// 示例：基础用法
//
// 演示 RegisterBean / Provide / aware 标签注入 / GetBean 的最简流程。
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct {
	Prefix string
}

// UserService 通过 aware 标签声明依赖，容器自动注入。
type UserService struct {
	DB *DB `aware:"db"`
}

func main() {
	// 注册一个已实例化的 bean（指针）
	di.RegisterNamedBean("db", &DB{Prefix: "tbl_"})

	// 注册结构体原型，由容器实例化并注入
	di.Provide(UserService{})

	// 加载容器（触发实例化、依赖注入、生命周期）
	di.Load()

	// 按名获取
	svc, ok := di.GetBean("userService")
	if !ok {
		panic("userService not found")
	}
	fmt.Printf("userService.DB.Prefix = %q\n", svc.(*UserService).DB.Prefix)
}
