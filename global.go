package di

import (
	"context"
	"os"
	"strings"
)

var g DI

func init() {
	g = New()
}

func RegisterBean(bean interface{}) DI {
	return g.RegisterBean(bean)
}

func RegisterNamedBean(name string, bean interface{}) DI {
	return g.RegisterNamedBean(name, bean)
}

func Provide(prototype interface{}) DI {
	return g.Provide(prototype)
}

func ProvideNamedBean(beanName string, prototype interface{}) DI {
	return g.ProvideNamedBean(beanName, prototype)
}

func GetBean(beanName string) (bean interface{}, ok bool) {
	return g.GetBean(beanName)
}

func GetByType(beanType interface{}) (bean interface{}, ok bool) {
	return g.GetByType(beanType)
}

func NewBean(beanType interface{}) (bean interface{}) {
	return g.NewBean(beanType)
}

func NewBeanByName(beanName string) (bean interface{}) {
	return g.NewBeanByName(beanName)
}

func UseValueStore(v ValueStore) DI {
	g.UseValueStore(v)
	return g
}

func Property() ValueStore {
	return g.Property()
}

func SetDefaultProperty(key string, value interface{}) DI {
	return g.SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]interface{}) DI {
	return g.SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value interface{}) DI {
	return g.SetProperty(key, value)
}

func SetPropertyMap(properties map[string]interface{}) DI {
	return g.SetPropertyMap(properties)
}

func AutoMigrateEnv() {
	envMap := LoadEnvironment(strings.NewReplacer("_", "."), false)
	SetPropertyMap(envMap)
}

func LoadEnvironment(replacer *strings.Replacer, trimPrefix bool, prefix ...string) map[string]interface{} {
	environ := os.Environ()
	envMap := make(map[string]interface{}, len(environ))
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

func Load() {
	g.Load()
}

func Serve(ctx context.Context) {
	g.Serve(ctx)
}

func LoadAndServ(ctx context.Context) {
	g.Load()
	g.Serve(ctx)
}
