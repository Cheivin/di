package di

import "reflect"

type ValueStore interface {
	SetDefault(key string, value interface{})

	Set(key string, value interface{})

	Get(key string) (val interface{})

	GetAll() map[string]interface{}
}

func (container *di) UseValueStore(v ValueStore) DI {
	container.valueStore = v
	return container
}

func (container *di) Property() ValueStore {
	return container.valueStore
}

func (container *di) SetDefaultProperty(key string, value interface{}) DI {
	container.valueStore.SetDefault(key, value)
	return container
}

func (container *di) SetDefaultPropertyMap(properties map[string]interface{}) DI {
	for key, value := range properties {
		container.valueStore.SetDefault(key, value)
	}
	return container
}

func (container *di) SetProperty(key string, value interface{}) DI {
	container.valueStore.Set(key, value)
	return container
}

func (container *di) SetPropertyMap(properties map[string]interface{}) DI {
	for key, value := range properties {
		container.valueStore.Set(key, value)
	}
	return container
}

func (container *di) GetProperty(key string) interface{} {
	return container.valueStore.Get(key)
}

func (container *di) LoadProperties(prefix string, propertyType interface{}) interface{} {
	prototype := reflect.Indirect(reflect.ValueOf(propertyType)).Type()
	def := container.getValueDefinition(prototype)
	bean := reflect.New(def.Type)
	container.wireValue(bean.Elem(), def, prefix)
	return bean.Elem().Interface()
}
