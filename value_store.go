package di

import (
	"os"
	"reflect"
	"strings"
)

// ValueStore 是配置存储的抽象接口，默认实现为 van.Van。
// di 通过它读写配置项，配合 value 标签实现配置注入。
type ValueStore interface {
	// SetDefault 设置默认值（低优先级，被 Set 覆盖）
	SetDefault(key string, value any)
	// Set 设置值（高优先级）
	Set(key string, value any)
	// Get 获取值（优先 Set，其次 SetDefault）
	Get(key string) (val any)
	// GetAll 获取所有配置的合并结果
	GetAll() map[string]any
}

// UseValueStore 替换配置存储实现。必须在 Load 前调用。
func (container *di) UseValueStore(v ValueStore) DI {
	container.valueStore = v
	return container
}

// Property 返回当前配置存储。
func (container *di) Property() ValueStore {
	return container.valueStore
}

// SetDefaultProperty 设置默认配置项（低优先级）。
// 注意：valueStore 的读写未加锁，配置应在 Load 前设置，避免与注入阶段的读取并发。
func (container *di) SetDefaultProperty(key string, value any) DI {
	container.valueStore.SetDefault(key, value)
	return container
}

// SetDefaultPropertyMap 批量设置默认配置项。
func (container *di) SetDefaultPropertyMap(properties map[string]any) DI {
	for key, value := range properties {
		container.valueStore.SetDefault(key, value)
	}
	return container
}

// SetProperty 设置配置项（高优先级，覆盖 SetDefaultProperty）。
func (container *di) SetProperty(key string, value any) DI {
	container.valueStore.Set(key, value)
	return container
}

// SetPropertyMap 批量设置配置项。
func (container *di) SetPropertyMap(properties map[string]any) DI {
	for key, value := range properties {
		container.valueStore.Set(key, value)
	}
	return container
}

// GetProperty 获取配置项值。
func (container *di) GetProperty(key string) any {
	return container.valueStore.Get(key)
}

// LoadProperties 将配置项按 prefix 前缀加载到 propertyType 结构体，
// 返回新构造并注入完成的实例（不回填传入的 propertyType，也不注册为 bean）。
func (container *di) LoadProperties(prefix string, propertyType any) any {
	prototype := reflect.Indirect(reflect.ValueOf(propertyType)).Type()
	def := container.getValueDefinition(prototype)
	bean := reflect.New(def.Type)
	container.wireValue(bean.Elem(), def, prefix)
	return bean.Elem().Interface()
}

// AutoMigrateEnv 读取所有环境变量并作为配置项注入。
// key 中的下划线 _ 转换为点号 .（如 APP_PORT → app.port）。
func (container *di) AutoMigrateEnv() DI {
	container.SetPropertyMap(LoadEnvironment(strings.NewReplacer("_", "."), false))
	return container
}

// LoadEnvironment 读取环境变量并返回 map。
// replacer 对 key 做替换（如 _ → .）；trimPrefix 为 true 时去掉 prefix 命中的前缀。
// prefix 为空时读取全部环境变量。
//
// 这是 AutoMigrateEnv 的底层构建块，也可独立使用：
//
//	envMap := di.LoadEnvironment(strings.NewReplacer("_", "."), true, "APP_")
//	c.SetPropertyMap(envMap)
func LoadEnvironment(replacer *strings.Replacer, trimPrefix bool, prefix ...string) map[string]any {
	environ := os.Environ()
	envMap := make(map[string]any, len(environ))
	for _, env := range environ {
		kv := strings.SplitN(env, "=", 2)
		if ok, pfx := hasPrefix(kv[0], prefix); !ok {
			continue
		} else if trimPrefix {
			kv[0] = strings.TrimPrefix(kv[0], pfx)
		}
		var property string
		if replacer != nil {
			property = replacer.Replace(kv[0])
		} else {
			property = kv[0]
		}
		envMap[property] = kv[1]
	}
	return envMap
}
