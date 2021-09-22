# di

`di`是一个简易版本的Go依赖注入实现

- [di](#di)
    * [安装](#安装)
    * [快速使用](#快速使用)
    * [函数](#函数)
        + [New](#new)
        + [RegisterBean](#registerbean)
        + [RegisterNamedBean](#registernamedbean)
        + [Provide](#provide)
        + [ProvideNamedBean](#ProvideNamedBean)
        + [GetBean](#getbean)
        + [Property](#Property)
        + [UseValueStore](#ValueStore)
        + [Serve](#Serve)
        + [LoadAndServe](#LoadAndServe)
    * [标签](#标签)
        + [aware](#aware)
        + [value](#value)
    * [接口](#接口)
        + [BeanConstruct](#beanconstruct)
        + [BeanName](#beanName)
        + [PreInitialize](#preinitialize)
        + [AfterPropertiesSet](#afterpropertiesset)
        + [Initialized](#initialized)
        + [Disposable](#disposable)
    * [配置项管理器](#配置项管理器)
        + [接口方法](#接口方法)
        + [内置管理器van](#内置管理器van)
    * [其他](#其他)
        + [UnsafeMode不安全模式](#unsafemode不安全模式)
        + [beanName生成策略](#beanname生成策略)

---

## 安装

```
go get github.com/cheivin/di
```

## 快速使用

```go
package main

import (
	"fmt"
	"github.com/cheivin/di"
	"log"
)

type (
	DB struct {
		Prefix string
	}

	UserDao struct {
		Db          *DB `aware:"db"`
		TableName   string
		DefaultAge  int    `value:"base.user.age"`
		DefaultName string `value:"base.user.name"`
		DefaultType uint8  `value:"base.user.type"`
	}

	WalletDao struct {
		Db        *DB `aware:"db"`
		TableName string
	}

	OrderRepository interface {
		TableName() string
	}

	OrderDao struct {
		Db *DB `aware:"db"`
	}

	UserService struct {
		UserDao  *UserDao        `aware:"user"`
		Wallet   *WalletDao      `aware:""`
		OrderDao OrderRepository `aware:"orderDao"`
	}
)

func (o *OrderDao) TableName() string {
	return o.Db.Prefix + "order"
}

func (u UserService) PreInitialize() {
	fmt.Println("依赖注入", "UserService")
}

func (u *UserDao) AfterPropertiesSet() {
	fmt.Println("装载完成", "UserDao")
	u.TableName = "user"
}

func (w *WalletDao) Initialized() {
	fmt.Println("加载完成", "WalletDao")
	w.TableName = "wallet"
}

func (o *OrderDao) BeanConstruct() {
	fmt.Println("构造实例", "OrderDao")
}

func (u *UserService) GetUserTable() string {
	return u.UserDao.Db.Prefix + u.UserDao.TableName
}

func (u *UserService) GetUserDefault() map[string]interface{} {
	return map[string]interface{}{
		"age":  u.UserDao.DefaultAge,
		"name": u.UserDao.DefaultName,
		"type": u.UserDao.DefaultType,
	}
}

func (u *UserService) GetWalletTable() string {
	return u.Wallet.Db.Prefix + u.Wallet.TableName
}

func (u *UserService) GetOrderTable() string {
	return u.OrderDao.TableName()
}

func main() {
	di.RegisterNamedBean("db", &DB{Prefix: "test_"}).
		ProvideNamedBean("user", UserDao{}).
		Provide(WalletDao{}).
		Provide(OrderDao{}).
		Provide(UserService{}).
		SetDefaultProperty("base.user.name", "新用户").
		SetProperty("base.user.age", 25).
		SetProperty("base.user.name", "新注册用户").
		SetProperty("base.user.type", "8").
		Load()

	bean, ok := di.GetBean("userService")
	if ok {
		log.Println(bean.(*UserService).GetUserTable())
		log.Println(bean.(*UserService).GetWalletTable())
		log.Println(bean.(*UserService).GetOrderTable())
		log.Println(bean.(*UserService).GetUserDefault())
	}
}
```

`Load()`函数表示开始自动装配

## 函数

### New

创建一个DI容器

```
container:=di.New()
```

### RegisterBean

手动注册一个bean对象，并根据类型自动确定beanName

- 注：手动注册的bean，不会触发依赖注入
- 可以动态注册bean，即在`Load()`方法执行后也可以使用

#### 参数

|  参数名   | 说明  |
|  ----  | ----  |
| bean  | 手动注册bean的指针 |

```go
package main

import "github.com/cheivin/di"

type AService struct {
	Name string
}

func main() {
	a := AService{Name: "a"}
	di.RegisterBean(&a)
}
```

### RegisterNamedBean

手动注册一个bean对象，并指定其beanName

- 注：手动注册的bean，不会触发依赖注入
- 可以动态注册bean，即在`Load()`方法执行后也可以使用

#### 参数

|  参数名   | 说明  |
|  ----  | ----  |
| beanName  | 指定的beanName |
| bean  | 手动注册bean的指针 |

```go
package main

import "github.com/cheivin/di"

type AService struct {
	Name string
}

func main() {
	a := AService{Name: "a"}
	di.RegisterNamedBean("aService", &a)
}
```

### Provide

提供一个bean结构体对象或其指针，容器将根据结构体自动创建bean并注入相关依赖

- 注：只能在`Load()`方法前使用

#### 参数

|  参数名   | 说明  |
|  ----  | ----  |
| prototype  | bean结构体对象 |

```go
package main

import "github.com/cheivin/di"

type (
	AService struct {
		Name string
	}
	BService struct {
	}
)

func main() {
	di.Provide(AService{}).
		Provide(&BService{})
}
```

### ProvideNamedBean

提供一个bean结构体对象或其指针，容器将使用指定名称根据结构体自动创建bean并注入相关依赖

- 注：只能在`Load()`方法前使用

#### 参数

|  参数名   | 说明  |
|  ----  | ----  |
| beanName  | 指定的beanName |
| prototype  | bean结构体对象 |

```go
package main

import "github.com/cheivin/di"

type AService struct {
	Name string
}

func main() {
	di.ProvideNamedBean("aService", AService{}).
		ProvideNamedBean("serviceA", &AService{})
}
```

### GetBean

根据beanName获取bean对象

#### 参数

|  参数名   | 说明  |
|  ----  | ----  |
| beanName  | 指定的beanName |

#### 返回值

|  返回值   | 说明  |
|  ----  | ----  |
| bean  | 获取到的bean对象 |
| ok  | 是否获取成功 |

```go
package main

import (
	"fmt"
	"github.com/cheivin/di"
)

type AService struct {
	Name string
}

func main() {
	di.ProvideNamedBean("aService", AService{}).
		Load()
	aService, ok := di.GetBean("aService")
	fmt.Println(aService, ok) // ok=true
	bService, ok := di.GetBean("bService")
	fmt.Println(bService, ok) // bService=nil,ok=false
}
```

### Property

返回配置项管理器，具体查看[配置项管理器](#配置项管理器)

### UseValueStore

设置配置项管理器，该管理器需实现`ValueStore`接口，具体查看[配置项管理器](#配置项管理器)

### Serve

在`Load()`方法后使用。用于监听context的退出。

如下，可实现程序优雅退出

```go
package main

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"os/signal"
	"syscall"
)

type AService struct {
	Name string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	di.Load()
	fmt.Println("容器加载完成")
	di.Serve(ctx)
	fmt.Println("程序退出")
}
```

### LoadAndServe

`Load()`和`Serve()`的合并命令。

## 标签

### aware

`DI`使用`aware`作为标记依赖注入的Tag

- Tag的完整格式为 `aware:"beanName"`
- Tag标记可以为`结构体指针`或`接口`，但不支持`基本数据类型`和`结构体`、`方法`
- 『**推荐**』如果Tag不传入任何值，即`aware:""`，则会根据[beanName生成策略](#beanname生成策略)得到beanName注入

```go
package main

import "github.com/cheivin/di"

type (
	AService          struct{}
	BService          struct{}
	CServiceInterface interface {
		Method()
	}
	CService struct {
	}
	BeanType struct {
		A *AService          `aware:"aService"` // 指定beanName为aService
		B *BService          `aware:""`         // 自动生成beanName为bService
		C *CServiceInterface `aware:"c"`        // 注入CService
	}
)

func (*CService) Method() {

}

func main() {
	di.ProvideNamedBean("c", CService{}).
		Provide(AService{}).
		Provide(BService{}).
		Provide(BeanType{}).
		Load()
}
```

### value

`DI`使用`value`作为标记配置项注入的Tag

- Tag的完整格式为 `value:"property"`
- Tag标记可以为以下`基本数据类型`：
    - `string`
    - `bool`
    - `int`
    - `int64`
    - `int32`
    - `int16`
    - `int8`
    - `uint`
    - `uint64`
    - `uint32`
    - `uint16`
    - `uint8`
    - `float64`
    - `float32`

```go
package main

import "github.com/cheivin/di"

type (
	BeanType struct {
		A string  `value:"a"`         // a
		B int     `value:"prop.b"`    // 1
		C float64 `value:"propMap.c"` // 1.0
	}
)

func main() {
	di.Provide(BeanType{}).
		SetProperty("a", "a").
		SetProperty("prop.b", 1).
		SetProperty("propMap", map[string]interface{}{
			"c": 1.0,
		}).
		Load()
}
```

## 接口

`DI`提供了以下接口，便于在依赖注入过程中处理部分特殊逻辑。

### BeanConstruct

`BeanConstruct()`在目标bean对象创建时触发。

- 此时仅反射创建对象，并未开始注入依赖属性。

### BeanName

`BeanName()`根据该方法返回值设定beanName。。

- 此时仅反射创建对象，并未开始注入依赖属性。

### PreInitialize

`PreInitialize()`在目标bean对象开始注入依赖时触发。

- 此时`DI`中的bean实例仅创建完成，但并未全部完成依赖注入。
- 其依赖的对象的依赖并不一定完成了注入。
- 此方法执行后，随即进行该bean的依赖注入。

### AfterPropertiesSet

`AfterPropertiesSet()`在目标bean对象完成注入依赖时触发。

- 此时`DI`中的bean实例仅创建完成，但并未全部完成依赖注入。
- 其依赖的对象的依赖并不一定完成了注入。
- 此方法之后，才可以通过`GetBean(当前beanName)`方法拿到当前bean

### Initialized

`Initialized()`在bean依赖注入完成后执行，可以理解为`DI`加载完成的通知事件。

- 此时`DI`中的bean均已完成依赖注入。
- 该接口用于bean依赖注入后，完成一些后置操作。

### Disposable

`Destroy()`在bean注销时执行。

- 当`DI`退出时，所有bean都会触发该方法。

## 配置项管理器

配置项管理器用于对`key-value`形式配置项进行管理。一个配置项管理器需实现`ValueStore`接口：

### 接口方法

- `SetDefault(key string, value interface{})` 设置默认配置值

- `Set(key string, value interface{})` 设置配置值

- `Get(key string) (val interface{})` 获取配置值

- `GetAll() map[string]interface{}` 获取所有配置

### 内置管理器van

`DI`提供了`van`作为默认的配置管理器。

- 可以独立使用，通过`van.New()`得到实例
- 支持设置形如`xxx.xxx.xxx`格式的配置
- 支持以`map[string]interface{}`方式设置多层级的配置项

```go
package main

import (
	"fmt"
	"github.com/cheivin/di/van"
)

func main() {
	store := van.New()
	store.SetDefault("a.b.c", "abc")
	store.SetDefault("a.b.d", "d")
	store.Set("a.b.c", "override")
	store.Set("a.b.e", "e")
	store.Set("a.b", map[string]interface{}{
		"x": 1,
		"y": 2,
		"z": map[string]interface{}{
			"n": 3,
		},
	})
	store.Set("a.b.z.m", 4)

	fmt.Println(store.Get("a.b.c"))   // override
	fmt.Println(store.Get("a.b.d"))   // d
	fmt.Println(store.Get("a.b.e"))   // e
	fmt.Println(store.Get("a.b.x"))   // 1
	fmt.Println(store.Get("a.b.z.n")) // 3
}
```

## 其他

### UnsafeMode不安全模式

golang通过大小写区分访问权限，私有属性默认无法通过反射修改完成注入

因此`DI`提供了不安全模式，通过`unsafe.Pointer`达到对私有属性的修改注入

```go
package main

import (
	"github.com/cheivin/di"
)

type (
	Dao struct {
	}
	AService struct {
		dao  *Dao   `aware:"dao"`
		name string `value:"name"`
	}
)

func main() {
	di.Provide(Dao{}).
		Provide(AService{}).
		SetProperty("name", "name").
		UnsafeMode(true). // 不安全模式
		Load()
}
```

### beanName生成策略

- 如果实现`BeanName`接口，则将把`BeanName`接口方法返回值作为beanName
- 否则将根据结构体名称，把第一个转换字母小写得到beanName
  - eg:
    - `AService` => `aService`
    - `DB` => `dB`
    - `api` => `api`