package di

import (
	"testing"
)

// 测试用的 Handler 接口和实现
type handler interface {
	Handle() string
}

// 两个不同的实现类型
type alphaHandler struct{}

func (alphaHandler) Handle() string    { return "alpha" }
func (alphaHandler) BeanName() string  { return "alpha" }

type betaHandler struct{}

func (betaHandler) Handle() string     { return "beta" }
func (betaHandler) BeanName() string   { return "beta" }

// 服务：通过 slice 收集所有 handler
type sliceOwner struct {
	Handlers []handler `aware:""`
}

// 服务：通过 map 收集所有 handler
type mapOwner struct {
	ByName map[string]handler `aware:""`
}

// 服务：slice 收集具体指针类型（非接口）
type repo interface{ Name() string }
type mysqlRepo struct{}

func (*mysqlRepo) Name() string  { return "mysql" }
func (*mysqlRepo) BeanName() string { return "mysqlRepo" }
type pgRepo struct{}

func (*pgRepo) Name() string  { return "pg" }
func (*pgRepo) BeanName() string { return "pgRepo" }

type repoOwner struct {
	Repos []repo `aware:""`
}

func TestSliceInject_InterfaceImpls(t *testing.T) {
	c := New()
	c.Provide(alphaHandler{})
	c.Provide(betaHandler{})
	c.Provide(sliceOwner{})
	c.Load()

	bean, ok := c.GetBean("sliceOwner")
	if !ok {
		t.Fatal("expected sliceOwner bean")
	}
	so := bean.(*sliceOwner)
	if len(so.Handlers) != 2 {
		t.Fatalf("want 2 handlers, got %d", len(so.Handlers))
	}
	// 按注册顺序
	if so.Handlers[0].Handle() != "alpha" {
		t.Errorf("handler[0] want alpha, got %s", so.Handlers[0].Handle())
	}
	if so.Handlers[1].Handle() != "beta" {
		t.Errorf("handler[1] want beta, got %s", so.Handlers[1].Handle())
	}
}

func TestMapInject_InterfaceImpls(t *testing.T) {
	c := New()
	c.Provide(alphaHandler{})
	c.Provide(betaHandler{})
	c.Provide(mapOwner{})
	c.Load()

	bean, ok := c.GetBean("mapOwner")
	if !ok {
		t.Fatal("expected mapOwner bean")
	}
	mo := bean.(*mapOwner)
	if len(mo.ByName) != 2 {
		t.Fatalf("want 2 entries, got %d", len(mo.ByName))
	}
	if h, exists := mo.ByName["alpha"]; !exists || h.Handle() != "alpha" {
		t.Errorf("map[alpha] missing or wrong: %v", mo.ByName["alpha"])
	}
	if h, exists := mo.ByName["beta"]; !exists || h.Handle() != "beta" {
		t.Errorf("map[beta] missing or wrong: %v", mo.ByName["beta"])
	}
}

func TestSliceInject_InterfacePtrImpls(t *testing.T) {
	c := New()
	c.RegisterBean(&mysqlRepo{})
	c.RegisterBean(&pgRepo{})
	c.Provide(repoOwner{})
	c.Load()

	bean, ok := c.GetBean("repoOwner")
	if !ok {
		t.Fatal("expected repoOwner bean")
	}
	ro := bean.(*repoOwner)
	if len(ro.Repos) != 2 {
		t.Fatalf("want 2 repos, got %d", len(ro.Repos))
	}
}

// TestSliceInject_EmptyCandidates 没有候选时注入空 slice（不报错）
func TestSliceInject_EmptyCandidates(t *testing.T) {
	type emptyOwner struct {
		Handlers []handler `aware:""`
	}
	c := New()
	c.Provide(emptyOwner{})
	c.Load()

	bean, ok := c.GetBean("emptyOwner")
	if !ok {
		t.Fatal("expected emptyOwner bean")
	}
	eo := bean.(*emptyOwner)
	if eo.Handlers == nil || len(eo.Handlers) != 0 {
		t.Fatalf("want non-nil empty slice, got %v", eo.Handlers)
	}
}
