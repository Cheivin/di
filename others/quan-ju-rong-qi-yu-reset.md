---
description: 全局容器懒初始化与 Reset 测试隔离（v0.4.0 新增）。
---

# 全局容器与 Reset

## 懒初始化

全局容器**不在 import 时创建**，首次调用全局函数时才创建：

```go
import "github.com/cheivin/di"
// 此时容器为 nil，import 零开销

func main() {
	di.RegisterBean(&DB{}) // 首次调用，创建容器
}
```

## Reset 重置

`di.Reset()` 将全局容器置为未初始化，下次调用时重新创建。主要用于测试隔离：

```go
func TestA(t *testing.T) {
	defer di.Reset() // 测试结束清理
	di.RegisterBean(&DB{})
	di.Load()
}

func TestB(t *testing.T) {
	defer di.Reset()
	// 全新容器，不受 TestA 影响
}
```

> ⚠️ `Reset()` 仅用于测试，生产代码不应调用。
