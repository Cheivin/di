module github.com/cheivin/di

go 1.25

// v0.4.0 和 v0.5.0 默认开启循环依赖检测，但 di 的两阶段设计天然支持
// 指针循环依赖，默认检测属于回归 bug。请升级到 v0.5.1 及以上。
retract (
	v0.4.0
	v0.5.0
)
