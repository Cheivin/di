---
layout: default
title: 安装
nav_order: 3
parent: 入门
---

# 安装

## 版本要求

- **Go 1.25+**（v0.4.0 起）

## 安装

```bash
go get github.com/cheivin/di@latest
```

## 验证

```go
package main

import (
	"fmt"

	"github.com/cheivin/di"
)

type DB struct{}

type Service struct {
	DB *DB `aware:""`
}

func main() {
	di.RegisterBean(&DB{})
	di.Provide(Service{})
	di.Load()

	svc, _ := di.GetBean("service")
	fmt.Println(svc.(*Service).DB != nil) // true
}
```

## 版本说明

di 遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。在 1.0 之前（0.x），minor 版本升级可能包含不兼容变更，详见 [CHANGELOG](https://github.com/cheivin/di/blob/main/CHANGELOG.md)。
