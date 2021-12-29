---
description: 默认的配置管理器
---

# 内置管理器van

### 说明

* 可以独立使用，通过 `van.New()` 得到实例
* 支持设置形如 `xx.xx.xx` 格式的配置项
* 支持以 `map[string]interface{}` 方式设置多层级的配置项

### 使用

```
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

