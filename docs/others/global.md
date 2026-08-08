---
layout: default
title: 全局容器与 Reset
nav_order: 6
parent: 其他
---

# 全局容器与 Reset

di 提供包级全局函数（`di.RegisterBean` 等），它们操作一个全局容器实例。

## 懒初始化（v0.4.0）

全局容器**不会在包导入时创建**，首次调用全局函数时才懒初始化：

```go
import "github.com/cheivin/di"

// 此时全局容器为 nil（import 零开销）

func main() {
	di.RegisterBean(&DB{}) // 首次调用，此时创建容器
}
```

这避免了"仅 import 但不使用"时的无谓初始化。

## 全局函数 vs 独立容器

| 方式 | 适用场景 |
|------|---------|
| 全局函数 `di.RegisterBean(...)` | 简单应用、快速原型 |
| 独立容器 `c := di.New(); c.RegisterBean(...)` | 需要多个容器、测试隔离、精细控制 |

全局函数等价于：

```go
var g DI // 懒初始化

func RegisterBean(bean any) DI {
	return container().RegisterBean(bean)
}

func container() DI {
	// 加锁，首次创建
	if g == nil { g = New() }
	return g
}
```

## Reset：重置全局容器

`di.Reset()` 将全局容器重置为未初始化状态，下次调用全局函数时重新创建：

```go
di.RegisterBean(&DB{})
di.Load()
// ... 使用

di.Reset() // 清空所有 bean 和配置，g = nil

// 下次调用时重新创建空容器
di.RegisterBean(&NewDB{})
```

### 测试隔离

全局容器有状态残留，不同测试之间会互相影响。用 `Reset()` 隔离：

```go
func TestA(t *testing.T) {
	defer di.Reset() // 测试结束清理
	di.RegisterBean(&DB{})
	di.Load()
	// ...
}

func TestB(t *testing.T) {
	defer di.Reset()
	// 这里的容器是全新的，不受 TestA 影响
}
```

> ⚠️ `Reset()` 仅用于测试，生产代码不应调用。
