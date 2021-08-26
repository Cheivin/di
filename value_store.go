package di

type ValueStore interface {
	SetDefault(key string, value interface{})

	Set(key string, value interface{})

	Get(key string) (val interface{})

	GetAll() map[string]interface{}
}

func (di *DI) UseValueStore(v ValueStore) *DI {
	di.valueStore = v
	return di
}

func (di *DI) Property() ValueStore {
	return di.valueStore
}

func (di *DI) SetDefaultProperty(key string, value interface{}) *DI {
	di.valueStore.SetDefault(key, value)
	return di
}

func (di *DI) SetDefaultPropertyMap(properties map[string]interface{}) *DI {
	for key, value := range properties {
		di.valueStore.SetDefault(key, value)
	}
	return di
}

func (di *DI) SetProperty(key string, value interface{}) *DI {
	di.valueStore.Set(key, value)
	return di
}

func (di *DI) SetPropertyMap(properties map[string]interface{}) *DI {
	for key, value := range properties {
		di.valueStore.Set(key, value)
	}
	return di
}
