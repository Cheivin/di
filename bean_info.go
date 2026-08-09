package di

import (
	"reflect"
	"slices"
	"strings"
)

// Dependency 描述 bean 的一个字段级依赖。
type Dependency struct {
	Field     string       // 字段名
	Name      string       // 注入的 bean 名（aware）或配置项 key（value）；slice/map 注入按类型收集，为空
	Type      reflect.Type // 字段类型
	Omitempty bool         // 是否可选（omitempty 标签）
}

// BeanDescription 描述 bean 定义（只读快照，供管理/诊断）。
// 仅原型（Provide）与工厂（ProvideFunc）bean 有定义；
// 直接注册的实例（RegisterBean）不经过定义解析，无描述信息（DescribeBean 返回 ok=false）。
type BeanDescription struct {
	Name         string       // bean 名称
	Type         reflect.Type // 类型（原型为结构体类型，工厂 bean 为返回指针类型）
	Factory      bool         // 是否为工厂模式
	Dependencies []Dependency // aware 依赖注入（按字段名排序）
	Values       []Dependency // value 配置注入（按字段名排序）
}

// DescribeBean 返回 bean 定义的只读描述；definition 不存在时返回 ok=false。
// 线程安全（读锁）。
func (container *di) DescribeBean(beanName string) (desc BeanDescription, ok bool) {
	container.mu.RLock()
	defer container.mu.RUnlock()
	def, ok := container.beanDefinitionMap[beanName]
	if !ok {
		return BeanDescription{}, false
	}
	deps := make([]Dependency, 0, len(def.awareMap))
	for field, aware := range def.awareMap {
		deps = append(deps, Dependency{
			Field:     field,
			Name:      aware.Name,
			Type:      aware.Type,
			Omitempty: aware.Omitempty,
		})
	}
	slices.SortFunc(deps, func(a, b Dependency) int { return strings.Compare(a.Field, b.Field) })
	values := make([]Dependency, 0, len(def.valueMap))
	for field, aware := range def.valueMap {
		values = append(values, Dependency{
			Field: field,
			Name:  aware.Name,
			Type:  aware.Type,
		})
	}
	slices.SortFunc(values, func(a, b Dependency) int { return strings.Compare(a.Field, b.Field) })
	return BeanDescription{
		Name:         def.Name,
		Type:         def.Type,
		Factory:      def.factory.IsValid(),
		Dependencies: deps,
		Values:       values,
	}, true
}

// GetBeanDependencies 返回 bean 依赖的其他 bean 名称列表（命名 aware 注入，按名称排序）。
// slice/map 按类型收集的注入（Name 为空）不包含在内；definition 不存在时返回 ok=false。
// 线程安全（读锁）。
func (container *di) GetBeanDependencies(beanName string) (deps []string, ok bool) {
	container.mu.RLock()
	defer container.mu.RUnlock()
	def, ok := container.beanDefinitionMap[beanName]
	if !ok {
		return nil, false
	}
	for _, aware := range def.awareMap {
		if aware.Name != "" {
			deps = append(deps, aware.Name)
		}
	}
	slices.Sort(deps)
	return deps, true
}
