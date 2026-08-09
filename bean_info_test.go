package di

import (
	"reflect"
	"testing"
)

type infoDepA struct{}

type infoDepB struct{}

type infoService struct {
	DepA *infoDepA `aware:""`
	DepB *infoDepB `aware:"dep-b"`
	Port int       `value:"server.port"`
	Name string    `value:"server.name"`
}

// TestDescribeBean 验证 bean 定义描述：aware/value 依赖按字段名排序，工厂标志正确。
func TestDescribeBean(t *testing.T) {
	container := New()
	container.Provide(infoService{})

	desc, ok := container.DescribeBean("infoService")
	if !ok {
		t.Fatal("DescribeBean should find infoService")
	}
	if desc.Name != "infoService" {
		t.Fatalf("desc.Name = %q, want infoService", desc.Name)
	}
	if desc.Factory {
		t.Fatal("infoService is prototype, Factory should be false")
	}
	if desc.Type != reflect.TypeFor[infoService]() {
		t.Fatalf("desc.Type = %v, want %v", desc.Type, reflect.TypeFor[infoService]())
	}

	// aware 依赖按字段名排序：DepA、DepB
	if len(desc.Dependencies) != 2 {
		t.Fatalf("Dependencies = %v, want 2 entries", desc.Dependencies)
	}
	wantDeps := []struct{ field, name string }{
		{"DepA", "infoDepA"}, // 空名称按类型推断
		{"DepB", "dep-b"},    // 显式名称
	}
	for i, want := range wantDeps {
		if desc.Dependencies[i].Field != want.field || desc.Dependencies[i].Name != want.name {
			t.Fatalf("Dependencies[%d] = %+v, want %+v", i, desc.Dependencies[i], want)
		}
	}

	// value 注入按字段名排序：Name、Port
	if len(desc.Values) != 2 {
		t.Fatalf("Values = %v, want 2 entries", desc.Values)
	}
	wantValues := []struct{ field, name string }{
		{"Name", "server.name"},
		{"Port", "server.port"},
	}
	for i, want := range wantValues {
		if desc.Values[i].Field != want.field || desc.Values[i].Name != want.name {
			t.Fatalf("Values[%d] = %+v, want %+v", i, desc.Values[i], want)
		}
	}

	// 不存在的 bean
	if _, ok := container.DescribeBean("notExist"); ok {
		t.Fatal("DescribeBean should return ok=false for missing bean")
	}
}

// TestDescribeBeanFactory 验证工厂 bean 的描述。
func TestDescribeBeanFactory(t *testing.T) {
	container := New()
	container.ProvideFunc(func() *infoDepA { return &infoDepA{} })

	desc, ok := container.DescribeBean("infoDepA")
	if !ok {
		t.Fatal("DescribeBean should find factory bean")
	}
	if !desc.Factory {
		t.Fatal("factory bean Factory should be true")
	}
}

// TestGetBeanDependencies 验证依赖列表按名称排序，slice/map 注入（无名称）不包含。
func TestGetBeanDependencies(t *testing.T) {
	container := New()
	type infoCollector struct {
		Services []*infoDepB `aware:"omitempty"` // 按类型收集，无名称
		DepA     *infoDepA   `aware:""`
		DepB     *infoDepB   `aware:"dep-b"`
	}
	container.Provide(infoCollector{})

	deps, ok := container.GetBeanDependencies("infoCollector")
	if !ok {
		t.Fatal("GetBeanDependencies should find infoCollector")
	}
	want := []string{"dep-b", "infoDepA"} // 按名称排序；slice 注入无名称不包含
	if len(deps) != len(want) {
		t.Fatalf("GetBeanDependencies = %v, want %v", deps, want)
	}
	for i := range want {
		if deps[i] != want[i] {
			t.Fatalf("GetBeanDependencies = %v, want %v", deps, want)
		}
	}

	// 不存在的 bean
	if _, ok := container.GetBeanDependencies("notExist"); ok {
		t.Fatal("GetBeanDependencies should return ok=false for missing bean")
	}
}
