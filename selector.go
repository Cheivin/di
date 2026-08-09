package di

import (
	"fmt"
	"reflect"
	"strings"
)

// BeanSelector 在一个接口/类型有多个候选实现时决定选中哪个。
// candidates 按 beanSort 注册顺序排列，至少 1 个元素。
// 返回选中索引；返回 error 会触发 Fatal。
type BeanSelector interface {
	Select(candidates []BeanWithName, targetType reflect.Type) (int, error)
}

// Primary 可被 bean 实现，声明自己为首选实现。
// PrimaryFirst 策略下，会优先选中实现了 Primary 的候选。
type Primary interface {
	IsPrimary() bool
}

// LastRegistered 取最后注册的候选。这是默认策略，与历史行为完全一致。
type LastRegistered struct{}

func (LastRegistered) Select(candidates []BeanWithName, _ reflect.Type) (int, error) {
	return len(candidates) - 1, nil
}

// FirstRegistered 取第一个注册的候选。
type FirstRegistered struct{}

func (FirstRegistered) Select(candidates []BeanWithName, _ reflect.Type) (int, error) {
	return 0, nil
}

// PrimaryFirst 优先返回实现了 Primary 的候选。
//   - 恰好一个 Primary：返回它
//   - 多个 Primary：返回 error（歧义）
//   - 无 Primary：回退到取最后一个（兼容默认）
type PrimaryFirst struct{}

func (PrimaryFirst) Select(candidates []BeanWithName, targetType reflect.Type) (int, error) {
	primaryIdx := -1
	for i, c := range candidates {
		if p, ok := c.Bean.(Primary); ok && p.IsPrimary() {
			if primaryIdx >= 0 {
				return -1, fmt.Errorf("%w: multiple @Primary beans for %s (%s and %s)",
					ErrBean, targetType.String(), candidates[primaryIdx].Name, c.Name)
			}
			primaryIdx = i
		}
	}
	if primaryIdx >= 0 {
		return primaryIdx, nil
	}
	return len(candidates) - 1, nil
}

// ErrorOnAmbiguous 严格模式：候选超过 1 个直接报错。
// 适用于希望显式命名注入、避免隐式依赖"最后注册"语义的项目。
type ErrorOnAmbiguous struct{}

func (ErrorOnAmbiguous) Select(candidates []BeanWithName, targetType reflect.Type) (int, error) {
	if len(candidates) > 1 {
		var names strings.Builder
		names.WriteString(candidates[0].Name)
		for i := 1; i < len(candidates); i++ {
			names.WriteString(", " + candidates[i].Name)
		}
		return -1, fmt.Errorf("%w: ambiguous beans for %s (%s); use named injection or @Primary",
			ErrBean, targetType.String(), names.String())
	}
	return 0, nil
}
