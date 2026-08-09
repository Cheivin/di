package di

import (
	"testing"
)

// TestGetBeanNames 验证按注册顺序返回所有 bean 名称（含实例、原型、工厂 bean）。
func TestGetBeanNames(t *testing.T) {
	container := New()

	// 未注册任何 bean 时返回空列表
	if names := container.GetBeanNames(); len(names) != 0 {
		t.Fatalf("empty container GetBeanNames = %v, want []", names)
	}

	type serviceA struct{}
	type serviceB struct{}
	type serviceC struct{}

	container.RegisterBean(&serviceA{})
	container.Provide(serviceB{})
	container.ProvideFunc(func(a *serviceA) *serviceC { return &serviceC{} })

	want := []string{"serviceA", "serviceB", "serviceC"}
	got := container.GetBeanNames()
	if len(got) != len(want) {
		t.Fatalf("GetBeanNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetBeanNames = %v, want %v", got, want)
		}
	}
}

// TestGetBeanNamesAfterLoad 验证 Load 后名称列表不变（bean 名由注册决定，与实例化无关）。
func TestGetBeanNamesAfterLoad(t *testing.T) {
	container := New()
	type serviceA struct{}
	container.RegisterBean(&serviceA{})
	before := container.GetBeanNames()
	container.Load()
	after := container.GetBeanNames()
	if len(before) != len(after) {
		t.Fatalf("GetBeanNames changed after Load: %v -> %v", before, after)
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("GetBeanNames changed after Load: %v -> %v", before, after)
		}
	}
}
