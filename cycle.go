package di

import (
	"errors"
	"reflect"
	"strings"
)

// ErrCircularDependency 循环依赖
var ErrCircularDependency = errors.New("circular dependency")

type circularError struct {
	chain []string
}

func (e *circularError) Error() string {
	if len(e.chain) == 0 {
		return ErrCircularDependency.Error()
	}
	var chain strings.Builder
	chain.WriteString(e.chain[0])
	for i := 1; i < len(e.chain); i++ {
		chain.WriteString(" -> " + e.chain[i])
	}
	return ErrCircularDependency.Error() + ": " + chain.String()
}

func (e *circularError) Unwrap() error {
	return ErrCircularDependency
}

// checkCircularDependency 基于 aware 依赖和 ProvideFunc 工厂入参做 DFS，发现环则返回环路径。
//
// 检测范围：
//   - aware 标签声明的依赖（命名依赖直接解析；接口依赖按类型匹配所有实现）
//   - ProvideFunc 工厂函数的入参依赖（按类型推断 beanName 或接口匹配）
//
// 仅在 Load() 启动期调用，无需加锁（调用方独占容器）。
func (container *di) checkCircularDependency() error {
	// 收集每个 bean 依赖的 beanName 集合（去重）
	deps := make(map[string][]string, len(container.beanDefinitionMap))
	for name, def := range container.beanDefinitionMap {
		seen := map[string]struct{}{}
		for _, a := range def.awareMap {
			for _, depName := range container.resolveDepNames(a, name) {
				if _, ok := seen[depName]; ok {
					continue
				}
				// 只关心能映射到已有 definition 的依赖；
				// 否则注入期会报 notfound，无需在此判环
				if _, exists := container.beanDefinitionMap[depName]; exists {
					seen[depName] = struct{}{}
					deps[name] = append(deps[name], depName)
				}
			}
		}
		// 工厂入参也算依赖（按类型推断名称）
		for _, argType := range def.factoryArgs {
			a := aware{Name: GetBeanName(argType), Type: argType}
			if argType.Kind() == reflect.Interface {
				a.Name = ""
				a.IsInterface = true
			}
			for _, depName := range container.resolveDepNames(a, name) {
				if _, ok := seen[depName]; ok {
					continue
				}
				if _, exists := container.beanDefinitionMap[depName]; exists {
					seen[depName] = struct{}{}
					deps[name] = append(deps[name], depName)
				}
			}
		}
	}

	const (
		white = 0 // 未访问
		gray  = 1 // 正在访问（在当前 DFS 栈中）
		black = 2 // 已完成
	)
	color := make(map[string]int, len(container.beanDefinitionMap))
	var path []string
	var found error

	// 按注册顺序遍历，错误信息更稳定
	var visit func(name string)
	visit = func(name string) {
		if found != nil {
			return
		}
		color[name] = gray
		path = append(path, name)
		for _, depName := range deps[name] {
			switch color[depName] {
			case gray:
				// 命中环：截取从 depName 开始到当前的路径，并闭合
				cycleStart := 0
				for i, p := range path {
					if p == depName {
						cycleStart = i
						break
					}
				}
				chain := append([]string{}, path[cycleStart:]...)
				chain = append(chain, depName)
				found = &circularError{chain: chain}
				return
			case white:
				visit(depName)
				if found != nil {
					return
				}
			}
		}
		path = path[:len(path)-1]
		color[name] = black
	}

	for _, name := range container.beanSort {
		// RegisterBean 注册的实例无 definition，跳过
		if _, ok := container.beanDefinitionMap[name]; !ok {
			continue
		}
		if color[name] == white {
			visit(name)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

// resolveDepNames 将一条 aware 信息解析为具体的 beanName 列表。
// 命名依赖：直接用 aware 名称。
// 无名称的接口依赖：按类型匹配所有实现。
//
// 仅在 checkCircularDependency（Load 启动期独占）内调用，直接读 map 不走 findBeanByName，
// 避免重入读锁。
func (container *di) resolveDepNames(a aware, ownerName string) []string {
	if a.Name != "" {
		return []string{a.Name}
	}
	if a.IsInterface {
		var names []string
		for _, depName := range container.beanSort {
			if depName == ownerName {
				continue
			}
			// 直接读 map（调用方处于启动期独占，无需加锁）
			bean, ok := container.beanMap[depName]
			if !ok {
				bean, ok = container.prototypeMap[depName]
			}
			if ok && bean != nil && reflect.TypeOf(bean).AssignableTo(a.Type) {
				names = append(names, depName)
			}
		}
		return names
	}
	return nil
}
