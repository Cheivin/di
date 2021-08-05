package di

var g *DI

func init() {
	g = New()
}

func RegisterBean(bean interface{}) *DI {
	return g.RegisterBean(bean)
}

func RegisterNamedBean(name string, bean interface{}) *DI {
	return g.RegisterNamedBean(name, bean)
}

func Provide(prototype interface{}) *DI {
	return g.Provide(prototype)
}

func ProvideWithBeanName(beanName string, prototype interface{}) *DI {
	return g.ProvideWithBeanName(beanName, prototype)
}

func GetBean(beanName string) (bean interface{}, ok bool) {
	return g.GetBean(beanName)
}

func UseValueStore(v ValueStore) *DI {
	g.UseValueStore(v)
	return g
}

func Property() ValueStore {
	return g.Property()
}

func SetDefaultProperty(key string, value interface{}) *DI {
	return g.SetDefaultProperty(key, value)
}

func SetProperty(key string, value interface{}) *DI {
	return g.SetProperty(key, value)
}

func Load() {
	if g.loaded {
		return
	}
	g.Load()
}
