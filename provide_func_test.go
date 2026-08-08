package di

import (
	"errors"
	"testing"
)

// 工厂依赖
type pfDB struct {
	Prefix string
}

func (*pfDB) BeanName() string { return "pfDB" }

// 工厂产出的服务
type pfService struct {
	db *pfDB
}

func (s *pfService) BeanName() string { return "pfService" }

// TestProvideFunc_Basic 工厂按入参类型注入依赖
func TestProvideFunc_Basic(t *testing.T) {
	c := New()
	c.RegisterBean(&pfDB{Prefix: "tbl_"})
	c.ProvideFunc(func(db *pfDB) *pfService {
		return &pfService{db: db}
	})
	c.Load()

	bean, ok := c.GetBean("pfService")
	if !ok {
		t.Fatal("expected pfService bean")
	}
	svc := bean.(*pfService)
	if svc.db == nil {
		t.Fatal("expected db injected")
	}
	if svc.db.Prefix != "tbl_" {
		t.Fatalf("want prefix tbl_, got %s", svc.db.Prefix)
	}
}

// TestProvideFunc_MultipleArgs 多入参
type pfLogger struct{}

func (*pfLogger) BeanName() string { return "pfLogger" }

type pfMultiService struct {
	db  *pfDB
	log *pfLogger
}

func (*pfMultiService) BeanName() string { return "pfMultiService" }

func TestProvideFunc_MultipleArgs(t *testing.T) {
	c := New()
	c.RegisterBean(&pfDB{Prefix: "multi_"})
	c.RegisterBean(&pfLogger{})
	c.ProvideFunc(func(db *pfDB, log *pfLogger) *pfMultiService {
		return &pfMultiService{db: db, log: log}
	})
	c.Load()

	bean, ok := c.GetBean("pfMultiService")
	if !ok {
		t.Fatal("expected pfMultiService bean")
	}
	svc := bean.(*pfMultiService)
	if svc.db == nil || svc.log == nil {
		t.Fatal("expected both deps injected")
	}
	if svc.db.Prefix != "multi_" {
		t.Fatalf("want prefix multi_, got %s", svc.db.Prefix)
	}
}

// TestProvideFunc_Lifecycle 工厂产物仍走生命周期
type pfLifecycle struct {
	constructed bool
	destroyed   bool
}

func (*pfLifecycle) BeanName() string { return "pfLifecycle" }
func (p *pfLifecycle) BeanConstruct() { p.constructed = true }
func (p *pfLifecycle) Destroy()       { p.destroyed = true }

func TestProvideFunc_Lifecycle(t *testing.T) {
	c := New()
	c.ProvideFunc(func() *pfLifecycle {
		return &pfLifecycle{}
	})
	c.Load()

	bean, ok := c.GetBean("pfLifecycle")
	if !ok {
		t.Fatal("expected pfLifecycle bean")
	}
	lc := bean.(*pfLifecycle)
	if !lc.constructed {
		t.Fatal("expected BeanConstruct called")
	}
}

// TestProvideFunc_InvalidInput 非函数 / 返回非指针 / 多返回值应报错
func TestProvideFunc_InvalidInput(t *testing.T) {
	expectFatalErr(t, func() {
		c := New()
		c.ProvideFunc("not a function")
	}, "expects a function")

	expectFatalErr(t, func() {
		c := New()
		c.ProvideFunc(func() {})
	}, "must return exactly one value")

	expectFatalErr(t, func() {
		c := New()
		type value struct{}
		c.ProvideFunc(func() value { return value{} })
	}, "must return a pointer")
}

// expectFatalErr 断言 fn 触发 Fatal panic（wrapping ErrBean）
func expectFatalErr(t *testing.T, fn func(), substr string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected fatal panic, got none")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("expected error, got %T: %v", r, r)
		}
		if !errors.Is(err, ErrBean) {
			t.Fatalf("want ErrBean, got %v", err)
		}
		if !contains(err.Error(), substr) {
			t.Fatalf("want message containing %q, got %q", substr, err.Error())
		}
	}()
	fn()
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
