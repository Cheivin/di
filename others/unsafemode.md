# UnsafeMode不安全模式

* `golang`通过大小写区分访问权限，未导出字段默认无法通过反射修改完成注入。
* 因此`DI`提供了不安全模式，通过`unsafe.Pointer`达到对私有属性的修改注入。

### 实例

```
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
