---
layout: default
title: 生命周期
nav_order: 7
parent: Bean 管理
---

# 生命周期

容器托管的所有 bean（无论是 `RegisterBean` 注册的实例、`Provide` 实例化的，还是 `ProvideFunc` 工厂产出的）都会按固定顺序触发生命周期回调。每个阶段都提供了带容器引用（`WithContainer`）和不带容器引用两个变体。

## 完整顺序

```
1. BeanConstruct          实例创建时（注入前，依赖字段为 nil）
2. PreInitialize          依赖注入前
3. 依赖注入               aware / value 字段被赋值
4. AfterPropertiesSet     注入完成
5. Initialized            所有 bean 加载完毕
6. Disposable             容器销毁时（按注册倒序）
```

每个阶段对应两个接口（plain 与 `WithContainer`）：

| 阶段 | 接口 | 带容器版本 |
|------|------|-----------|
| 1. 实例创建 | `BeanConstruct` | `BeanConstructWithContainer` |
| 2. 注入前 | `PreInitialize` | `PreInitializeWithContainer` |
| 4. 注入完成 | `AfterPropertiesSet` | `AfterPropertiesSetWithContainer` |
| 5. 加载完成 | `Initialized` | `InitializedWithContainer` |
| 6. 销毁 | `Disposable` | `DisposableWithContainer` |

## WithContainer 变体优先

当一个 bean 同时实现了 plain 版本和 `WithContainer` 版本时，**只调用 `WithContainer` 版本**，不会两个都调。带容器版本能拿到 `DI` 引用，可在回调里按需获取其他 bean。

```go
type AfterPropertiesSet interface {
    AfterPropertiesSet()
}
type AfterPropertiesSetWithContainer interface {
    AfterPropertiesSet(DI)
}
```

两者都未实现则该阶段跳过。

## 各阶段说明

### 1. BeanConstruct — 实例创建时

实例刚被创建（反射 `new` 或工厂调用返回）时触发。此时**依赖尚未注入**，所有 `aware` 字段都是 nil。适合做不依赖其他 bean 的初始化（如设置默认值）。

```go
func (s *Service) BeanConstruct() {
    s.pool = make([]Conn, 10) // 初始化自身状态
}
```

### 2. PreInitialize — 注入前

依赖注入即将开始前触发。可用于记录日志、做注入前的准备。

### 3. 依赖注入

容器扫描 `aware` / `value` 标签，按名称或类型查找依赖并赋值。这一步没有回调接口，由容器自动完成。

### 4. AfterPropertiesSet — 注入完成

所有依赖已就绪，可以安全使用。这是最常用的初始化钩子（建立连接、预热缓存、注册路由等）。

```go
func (s *Service) AfterPropertiesSet() {
    // s.DB 此时已注入，可安全使用
    s.warmup(s.DB)
}
```

### 5. Initialized — 所有 bean 加载完毕

容器内**所有** bean 都完成注入后才逐个触发。适合需要"全局就绪"后才能做的事情（如启动后台任务、打印就绪清单）。注意此时其他 bean 的 `Initialized` 可能尚未调用，不要假设顺序。

### 6. Disposable — 容器销毁时（倒序）

容器销毁（`Serve(ctx)` 收到 ctx 取消信号后，或调用内部 `destroyBeans`）时触发，**按注册倒序**执行：后注册的先销毁。用于释放连接、关闭文件、刷新缓冲。

```go
func (s *Service) Destroy() {
    s.DB.Close()
}
```

销毁由 `di.Serve(ctx)` 在收到信号时自动触发。直接调用 `di.Load()` 不会触发销毁，需配合 `Serve` 或手动管理。

## 示例：完整顺序

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

// 生命周期顺序：BeanConstruct → PreInitialize → 注入 → AfterPropertiesSet → Initialized → Destroy(倒序)

type DB struct{}

type Service struct {
	DB *DB `aware:""`
}

// BeanConstruct 实例创建时（注入前）
func (s *Service) BeanConstruct() {
	fmt.Println("[BeanConstruct] 依赖尚未注入，s.DB =", s.DB) // nil
}

// PreInitialize 注入前
func (s *Service) PreInitialize() {
	fmt.Println("[PreInitialize] 即将开始注入")
}

// AfterPropertiesSet 注入完成后
func (s *Service) AfterPropertiesSet() {
	fmt.Printf("[AfterPropertiesSet] 注入完成，s.DB = %T\n", s.DB) // *main.DB
}

// Initialized 所有 bean 加载完毕
func (s *Service) Initialized() {
	fmt.Println("[Initialized] 容器加载完成")
}

// Destroy 容器销毁时（倒序）
func (s *Service) Destroy() {
	fmt.Println("[Destroy] bean 销毁")
}

// 带 WithContainer 的变体：可获取容器引用
type AwareService struct {
	di.DI `aware:""`
}

func (a *AwareService) AfterPropertiesSet(container di.DI) {
	// 在回调里通过容器获取其他 bean
	db, ok := container.GetBean("dB")
	fmt.Printf("[AfterPropertiesSetWithContainer] 通过容器获取 db: %v\n", ok && db != nil)
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(Service{})
	di.Provide(AwareService{})
	di.Load()

	fmt.Println("\n=== 容器加载完成 ===")
	// 销毁由 di.Serve(ctx) 在收到信号时触发
}
```

输出顺序：

```
[BeanConstruct] 依赖尚未注入，s.DB = <nil>
[PreInitialize] 即将开始注入
[AfterPropertiesSet] 注入完成，s.DB = *main.DB
[AfterPropertiesSetWithContainer] 通过容器获取 db: true
[Initialized] 容器加载完成

=== 容器加载完成 ===
```

（`Destroy` 在 `Serve` 收到停止信号时才会打印，且按注册倒序：`awareService` 先于 `service` 销毁。）

## 注意事项

- **BeanConstruct 时依赖为 nil**：此时 `aware` 字段还没注入，不要在这里访问依赖。
- **WithContainer 优先**：同时实现两个变体只调带容器的那个。
- **销毁倒序**：与注册顺序相反，保证被依赖的 bean 后销毁。
- **NewBean 也走生命周期**：[NewBean](getbean) 创建的实例会触发 `BeanConstruct` → ... → `Initialized`，GC 回收时触发 `Destroy`。
- **匿名结构体字段不能实现这些接口**：容器会拒绝把生命周期接口"提升"到外层 bean，注册时报错。

## 相关

- [注册实例](registerbean) / [注册结构体](providebean) — 注册的 bean 自动进入生命周期
- [获取 bean](getbean) — `NewBean` 每次创建的实例也走完整生命周期
