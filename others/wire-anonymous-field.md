---
description: 依赖注入支持匿名字段的注入
---

# 注入匿名字段

### 限制

注入的匿名字段不能实现以下生命周期接口，因受golang语言特性其存在于注入的生命周期内，会造成重复执行。

* BeanConstruct
* PreInitialize
* AfterPropertiesSet
* Initialized
* Disposable

### 示例

```
package main

import (
	"github.com/cheivin/di"
)

type (
	Dao struct {
	}
	AService struct {
		*Dao `aware:""` // 匿名字段注入，根据BeanName接口指定
	}
)

func (Dao) BeanName() string {
	return "dao"
}

func main() {
	di.Provide(Dao{}).
		Provide(AService{}).
		Load()
}
```

