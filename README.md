# di

`di` 是一个简易版本的 Go 依赖注入实现

[文档地址](https://cheivin.gitbook.io/di/)

## 特性

* 支持手动注册 bean 实例
* 支持注册 bean 类型原型，由 DI 容器自动实例化并托管 bean 实例
* 支持构造函数注入（`ProvideFunc`），按入参类型自动注入依赖
* 支持根据名称、类型获取 DI 容器托管的 bean 实例
* 支持根据类型手动生成新的 bean 实例并返回
* 支持配置项注入并转换成对应的基本类型
* 支持匿名字段的 bean 注入
* 支持 slice / map 批量注入（收集同类型的所有 bean）
* 支持接口多实现的选择策略（`BeanSelector` / `Primary`）
* 支持完整生命周期管理（构造 / 注入前 / 注入后 / 初始化 / 销毁）
* 支持循环依赖检测，启动期快速定位问题
* 线程安全，支持运行期动态注册与并发读取

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct{}

type UserService struct {
	DB *DB `aware:""`
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(UserService{})
	di.Load()

	svc, _ := di.GetBean("userService")
	fmt.Println(svc.(*UserService).DB != nil) // true
}
```

## 特别鸣谢

[![JetBrains](https://raw.githubusercontent.com/kainonly/ngx-bit/main/resource/jetbrains.svg)](https://www.jetbrains.com/?from=cheivin)

感谢 [JetBrains](https://www.jetbrains.com/?from=cheivin) 提供的开源开发许可证。
