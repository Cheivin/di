package di

var g = New()

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

func Load() {
	if g.loaded {
		return
	}
	g.Load()
}
