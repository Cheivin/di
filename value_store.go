package di

import "reflect"

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
