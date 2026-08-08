---
layout: default
title: UnsafeMode 不安全模式
nav_order: 1
parent: 其他
---

# UnsafeMode 不安全模式

默认情况下，di 通过 `reflect.Value.Set` 注入字段，要求字段**可导出（首字母大写）**。

`UnsafeMode(true)` 允许注入**未导出（私有）字段**，通过 `unsafe.Pointer` 绕过导出限制。

## 用法

```go
type service struct {
	db *DB `aware:"db"` // 小写未导出字段
}

c := di.New()
c.UnsafeMode(true) // 开启不安全模式
c.RegisterNamedBean("db", &DB{})
c.Provide(service{})
c.Load()
```

## 注意事项

- 使用了 `unsafe` 包，可能影响后续 Go 版本兼容性
- 开启时容器会打印 warn 日志
- 仅在无法修改目标类型（如第三方库的结构体）时使用
- 生产代码建议优先把字段改为导出
