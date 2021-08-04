package di

type ValueStore interface {
	SetDefault(key string, value interface{})

	Set(key string, value interface{})

	Get(key string) (val interface{})

	GetAll() map[string]interface{}
}
