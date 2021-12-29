# 接口定义

### DI使用接口

在DI中可以通过以下接口方法使用配置管理器

```
type DI interface {
// ...

	// UseValueStore 设置配置项管理器
	UseValueStore(v ValueStore) DI
	
	// Property 获取配置项管理器
	Property() ValueStore

	// SetDefaultProperty 设置默认配置项
	SetDefaultProperty(key string, value interface{}) DI

	// SetDefaultPropertyMap 设置多个默认配置项
	SetDefaultPropertyMap(properties map[string]interface{}) DI

	// SetProperty 设置配置项
	SetProperty(key string, value interface{}) DI

	// SetPropertyMap 设置多个配置项
	SetPropertyMap(properties map[string]interface{}) DI
}
```

### 配置管理器定义

实现以下接口的结构体，可被设置为DI的配置项管理器

```
type ValueStore interface {
	// SetDefault 设置默认配置
	SetDefault(key string, value interface{})

	// Set 设置配置
	Set(key string, value interface{})

	// Get 获取配置
	Get(key string) (val interface{})

	// GetAll 获取所有配置
	GetAll() map[string]interface{}
}
```

