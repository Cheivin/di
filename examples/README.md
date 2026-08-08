# di 示例

每个子目录是一个独立的 Go module，演示 di 的一个特性。

## 运行方式

每个示例可独立运行（需在对应目录下）：

```bash
cd basic && go run .
```

或从本目录批量运行：

```bash
for d in */; do echo "=== $d ==="; (cd "$d" && go run .); done
```

## 示例清单

| 目录 | 演示特性 |
|------|---------|
| [basic](./basic) | 基础用法：RegisterBean / Provide / aware 注入 / GetBean |
| [provide-func](./provide-func) | 构造函数注入：ProvideFunc 工厂函数按入参类型注入 |
| [slice-inject](./slice-inject) | slice/map 批量注入：收集同接口的所有实现 |
| [selector](./selector) | 接口歧义策略：BeanSelector / Primary 多实现选择 |
| [lifecycle](./lifecycle) | 生命周期：BeanConstruct → PreInitialize → AfterPropertiesSet → Initialized → Destroy |
| [cycle-detection](./cycle-detection) | 循环依赖检测：Load 时自动发现并报错 |
| [config](./config) | 配置注入：value 标签 + van 类型转换（含 Duration 毫秒兜底） |

## 依赖说明

每个示例的 `go.mod` 通过 `replace github.com/cheivin/di => ../..` 指向本地 di 源码。
发布到独立仓库时，将 replace 删除并指定版本号即可。
