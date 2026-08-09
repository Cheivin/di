package di

import (
	"testing"
)

type hasTypeA struct{}
type hasTypeB struct{}

func (hasTypeA) M() {}

type hasTypeIface interface{ M() }

// TestHasBeanType 验证三种注册形式的类型查询：实例/原型/工厂 bean。
func TestHasBeanType(t *testing.T) {
	container := New()
	if container.HasBeanType(hasTypeA{}) {
		t.Fatal("empty container should not have hasTypeA")
	}

	// 注册实例
	container.RegisterBean(&hasTypeA{})
	if !container.HasBeanType(hasTypeA{}) {
		t.Fatal("registered instance should match hasTypeA")
	}
	if !container.HasBeanType((*hasTypeA)(nil)) {
		t.Fatal("registered instance should match (*hasTypeA)")
	}
	// 接口类型：*hasTypeA 实现 hasTypeIface
	if !container.HasBeanType((*hasTypeIface)(nil)) {
		t.Fatal("registered instance should match hasTypeIface")
	}

	// 注册原型
	container.Provide(hasTypeB{})
	if !container.HasBeanType(hasTypeB{}) {
		t.Fatal("provided prototype should match hasTypeB")
	}

	// 注册工厂 bean
	type hasTypeC struct{}
	container.ProvideFunc(func() *hasTypeC { return &hasTypeC{} })
	if !container.HasBeanType(hasTypeC{}) {
		t.Fatal("factory bean should match hasTypeC")
	}
	if !container.HasBeanType((*hasTypeC)(nil)) {
		t.Fatal("factory bean should match (*hasTypeC)")
	}

	// 未注册的类型
	type hasTypeOther interface{ Other() }
	if container.HasBeanType((*hasTypeOther)(nil)) {
		t.Fatal("unregistered type should not match")
	}
	if container.HasBeanType(nil) {
		t.Fatal("nil beanType should not match")
	}
}

// TestHasBeanTypeAfterLoad 验证 Load 后查询仍可用。
func TestHasBeanTypeAfterLoad(t *testing.T) {
	container := New()
	container.Provide(hasTypeA{})
	if !container.HasBeanType(hasTypeA{}) {
		t.Fatal("prototype should match before Load")
	}
	container.Load()
	if !container.HasBeanType(hasTypeA{}) {
		t.Fatal("prototype should match after Load")
	}
}
