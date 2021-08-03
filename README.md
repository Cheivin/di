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
        + [ProvideWithBeanName](#providewithbeanname)
        + [GetBean](#getbean)
    * [标签](#标签)
    * [接口](#接口)
        + [BeanConstruct](#beanconstruct)
        + [PreInitialize](#preinitialize)
        + [AfterPropertiesSet](#afterpropertiesset)
        + [Initialized](#initialized)
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
		Db        DB `aware:"db"`
		TableName string
	}

	WalletDao struct {
		Db DB `aware:"db"`
	}

	UserService struct {
		UerDao *UserDao `aware:"user"`
		Wallet *WalletDao
	}
)

func (u UserService) InitializeBean() {
	fmt.Println("依赖注入", "UserService")
}

func (u *UserDao) AfterPropertiesSet() {
	fmt.Println("装载完成", "UserDao")
	u.TableName = "user"
}

func (w WalletDao) BeanConstruct() {
	fmt.Println("构造对象", "WalletDao")
}

func (u *UserService) GetUserTable() string {
	return u.UerDao.Db.Prefix + u.UerDao.TableName
}

func main() {
	di.RegisterNamedBean("db", &DB{Prefix: "test_"}).
		Provide(WalletDao{}).
		ProvideWithBeanName("user", UserDao{}).
		Provide(UserService{}).
		Load()

	if bean, ok := di.GetBean("userService"); ok {
		log.Println(bean.(*UserService).GetUserTable())
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

- 注：手动注册的bean，不会触发依赖出入

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

- 注：手动注册的bean，不会触发依赖出入

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

### ProvideWithBeanName

提供一个bean结构体对象或其指针，容器将使用指定名称根据结构体自动创建bean并注入相关依赖

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
	di.ProvideWithBeanName("aService", AService{}).
		ProvideWithBeanName("serviceA", &AService{})
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
	di.ProvideWithBeanName("aService", AService{}).
		Load()
	aService, ok := di.GetBean("aService")
	fmt.Println(aService, ok) // ok=true
	bService, ok := di.GetBean("bService")
	fmt.Println(bService, ok) // bService=nil,ok=false
}
```

## 标签

`DI`使用`aware`作为标记依赖注入的Tag

- Tag的完整格式为 `aware:"beanName"`
- Tag标记的属性，可以为`结构`和`结构指针`，但不支持`基本数据类型`和`接口`、`方法`
- 『**不推荐**』如果Tag不传入任何值，即`aware:""`，则会根据字段结构名称自动生成beanName

```go
package main

type (
	AService struct{}
	BService struct{}
	CService struct{}
	BeanType struct {
		A *AService `aware:"aService"` // 单例
		B *BService `aware:""`         // 单例
		C BService  `aware:"aService"` // 多例
	}
)
```

## 接口

`DI`提供了以下接口，便于在依赖注入过程中处理部分特殊逻辑。

### BeanConstruct

`BeanConstruct()`在目标bean对象创建时触发。

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
		dao Dao `aware`
	}
)

func main() {
	di.Provide(Dao{}).
		Provide(AService{}).
		UnsafeMode(true). // 不安全模式
		Load()
}
```

### beanName生成策略

根据结构体名称，将第一个转换字母小写得到beanName

eg:

- `AService` => `aService`
- `DB` => `dB`
- `api` => `api`