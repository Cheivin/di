package di

import (
	"context"
	"os"
	"strings"
	"sync"
)

var (
	gMu sync.Mutex
	g   DI
)

// container 懒初始化并返回全局容器实例。
// 未调用任何全局函数前不创建容器，首次访问时创建。
func container() DI {
	gMu.Lock()
	defer gMu.Unlock()
	if g == nil {
		g = New()
	}
	return g
}

// Reset 将全局容器重置为未初始化状态（清空所有 bean 与配置，下次调用时懒创建）。
// 仅用于测试隔离（全局容器有状态残留），生产代码不应调用。
func Reset() {
	gMu.Lock()
	defer gMu.Unlock()
	g = nil
}

func RegisterBean(bean any) DI {
	return container().RegisterBean(bean)
}

func RegisterNamedBean(name string, bean any) DI {
	return container().RegisterNamedBean(name, bean)
}

func Provide(prototype any) DI {
	return container().Provide(prototype)
}

func ProvideNamedBean(beanName string, prototype any) DI {
	return container().ProvideNamedBean(beanName, prototype)
}

func ProvideFunc(fn any) DI {
	return container().ProvideFunc(fn)
}

func GetBean(beanName string) (bean any, ok bool) {
	return container().GetBean(beanName)
}

func GetByType(beanType any) (bean any, ok bool) {
	return container().GetByType(beanType)
}

func GetByTypeAll(beanType any) (beans []BeanWithName) {
	return container().GetByTypeAll(beanType)
}

func NewBean(beanType any) (bean any) {
	return container().NewBean(beanType)
}

func NewBeanByName(beanName string) (bean any) {
	return container().NewBeanByName(beanName)
}

func UseValueStore(v ValueStore) DI {
	c := container()
	c.UseValueStore(v)
	return c
}

func Property() ValueStore {
	return container().Property()
}

func SetDefaultProperty(key string, value any) DI {
	return container().SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]any) DI {
	return container().SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value any) DI {
	return container().SetProperty(key, value)
}

func SetPropertyMap(properties map[string]any) DI {
	return container().SetPropertyMap(properties)
}

func GetProperty(key string) any {
	return container().GetProperty(key)
}

func LoadProperties(prefix string, propertyType any) any {
	return container().LoadProperties(prefix, propertyType)
}

func AutoMigrateEnv() {
	envMap := LoadEnvironment(strings.NewReplacer("_", "."), false)
	SetPropertyMap(envMap)
}

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

func Load() {
	container().Load()
}

func Serve(ctx context.Context) {
	container().Serve(ctx)
}

func LoadAndServ(ctx context.Context) {
	container().Load()
	container().Serve(ctx)
}
