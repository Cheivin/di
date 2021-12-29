# 快速入门

### 要求

* Go 1.13及以上版本

### 安装

* 下载并安装DI：

```
go get -u github.com/cheivin/di
```

* 将DI引入代码中

```
import "github.com/cheivin/di"
```

### 开始

接下来，假设你有`Service`、`Dao`两个类，且`Service`依赖`Dao`

```
// definition.go
package main

type (
	Dao struct {
		Prefix string `value:"db.prefix"`
	}
	Service struct {
		Dao *Dao `aware:""`
	}
)

func (s Service) GetPrefix() string {
	return s.Dao.Prefix
}
```

在项目中配置并使用

```
// main.go
package main

import (
	"fmt"
	"github.com/cheivin/di"
)

func main() {
	di.SetProperty("db.prefix", "demo").
		Provide(Dao{}).
		Provide(Service{}).
		Load()
	bean, ok := di.GetByType(Service{})
	if !ok {
		panic("service 不存在")
	}
	service := bean.(*Service)
	fmt.Println("输出", service.GetPrefix())
}
```

然后运行代码，控制台会输出如下内容并退出：

```
[DI-INFO] : provide bean with name: dao
[DI-INFO] : provide bean with name: service
[DI-INFO] : wire value for bean dao(main.Dao)
[DI-INFO] : initialize bean dao(*main.Dao)
[DI-INFO] : initialize bean service(*main.Service)
[DI-INFO] : wire field for bean service(main.Service)
输出 demo
```

如需要使用`context`控制容器退出，则请使用 `Serve` 。以监听终端信号为例，修改后如下：

```
// main.go
package main

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"os/signal"
	"syscall"
)

func main() {
	di.SetProperty("db.prefix", "demo").
		Provide(Dao{}).
		Provide(Service{}).
		Load()
	bean, ok := di.GetByType(Service{})
	if !ok {
		panic("service 不存在")
	}
	service := bean.(*Service)
	fmt.Println("输出", service.GetPrefix())

	// 退出信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	di.Serve(ctx)
}
```

### 其他说明

* 使用包名di调用方法实际上是使用的全局容器实例。
* 可以使用`di.New()`手动创建多个容器。

