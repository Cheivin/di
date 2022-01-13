package di

import "context"

type DI interface {
	DebugMode(bool) DI

	Log(log Log) DI

	RegisterBean(bean interface{}) DI

	RegisterNamedBean(name string, bean interface{}) DI

	Provide(prototype interface{}) DI

	ProvideNamedBean(beanName string, prototype interface{}) DI

	GetBean(beanName string) (bean interface{}, ok bool)

	GetByType(beanType interface{}) (bean interface{}, ok bool)

	GetByTypeAll(beanType interface{}) (beans []BeanWithName)

	NewBean(beanType interface{}) (bean interface{})

	NewBeanByName(beanName string) (bean interface{})

	UseValueStore(v ValueStore) DI

	Property() ValueStore

	SetDefaultProperty(key string, value interface{}) DI

	SetDefaultPropertyMap(properties map[string]interface{}) DI

	SetProperty(key string, value interface{}) DI

	SetPropertyMap(properties map[string]interface{}) DI

	GetProperty(key string) interface{}

	LoadProperties(prefix string, propertyType interface{}) interface{}

	Load()

	Serve(ctx context.Context)

	Context() context.Context
}
