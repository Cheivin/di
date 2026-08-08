package di

import (
	"strings"
	"testing"
)

// ===== 生命周期顺序测试 =====

// lifecycleRecorder 记录回调调用顺序（包级实例，生命周期回调在注入前触发，不能依赖 aware 注入）
type lifecycleRecorder struct {
	events []string
}

var testRecorder = &lifecycleRecorder{}

func (r *lifecycleRecorder) add(e string) {
	r.events = append(r.events, e)
}

func (r *lifecycleRecorder) reset() {
	r.events = nil
}

type lifecycleBean struct {
	Rec *lifecycleRecorder `aware:""`
}

func (b *lifecycleBean) BeanConstruct() {
	testRecorder.add("construct")
}

func (b *lifecycleBean) PreInitialize() {
	testRecorder.add("pre-init")
}

func (b *lifecycleBean) AfterPropertiesSet() {
	testRecorder.add("after-props")
}

func (b *lifecycleBean) Initialized() {
	testRecorder.add("initialized")
}

func (b *lifecycleBean) Destroy() {
	testRecorder.add("destroy")
}

// TestLifecycle_Order 验证生命周期按 BeanConstruct → PreInitialize → AfterPropertiesSet → Initialized → Destroy 顺序
func TestLifecycle_Order(t *testing.T) {
	testRecorder.reset()
	c := New()
	c.RegisterBean(testRecorder)
	c.Provide(lifecycleBean{})
	c.Load()

	want := []string{"construct", "pre-init", "after-props", "initialized"}
	got := testRecorder.events
	if len(got) != len(want) {
		t.Fatalf("want %d events, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] want %q, got %q (all: %v)", i, want[i], got[i], got)
		}
	}
}

// ===== WithContainer 变体优先 =====

type containerAwareBean struct {
	Rec *lifecycleRecorder `aware:""`
}

func (b *containerAwareBean) BeanConstruct(di DI) {
	if di != nil {
		testRecorder.add("construct-with-container")
	}
}

func (b *containerAwareBean) PreInitialize(di DI) {
	if di != nil {
		testRecorder.add("pre-init-with-container")
	}
}

func (b *containerAwareBean) AfterPropertiesSet(di DI) {
	if di != nil {
		testRecorder.add("after-props-with-container")
	}
}

func (b *containerAwareBean) Initialized(di DI) {
	if di != nil {
		testRecorder.add("initialized-with-container")
	}
}

// TestLifecycle_WithContainerVariant 实现 WithContainer 变体时应优先调用带容器版本
func TestLifecycle_WithContainerVariant(t *testing.T) {
	testRecorder.reset()
	c := New()
	c.RegisterBean(testRecorder)
	c.Provide(containerAwareBean{})
	c.Load()

	// 只检查事件名含 -with-container
	for _, e := range testRecorder.events {
		if !strings.Contains(e, "with-container") {
			t.Fatalf("want with-container variant, got %q (all: %v)", e, testRecorder.events)
		}
	}
	if len(testRecorder.events) != 4 {
		t.Fatalf("want 4 lifecycle events, got %d: %v", len(testRecorder.events), testRecorder.events)
	}
}

// ===== value 标签注入 =====

type valueBean struct {
	Name     string  `value:"demo.name"`
	Port     int     `value:"demo.port"`
	Rate     float64 `value:"demo.rate"`
	Enabled  bool    `value:"demo.enabled"`
	Duration int64   `value:"demo.duration"` // int64 目标
}

// TestValueInjection 配置项通过 value 标签注入并类型转换
func TestValueInjection(t *testing.T) {
	c := New()
	c.SetDefaultProperty("demo.name", "demo")
	c.SetDefaultProperty("demo.port", 8080)
	c.SetDefaultProperty("demo.rate", "0.5") // 字符串转 float64
	c.SetDefaultProperty("demo.enabled", "true")
	c.SetDefaultProperty("demo.duration", "30000")
	c.Provide(valueBean{})
	c.Load()

	bean, ok := c.GetBean("valueBean")
	if !ok {
		t.Fatal("expected valueBean")
	}
	vb := bean.(*valueBean)
	if vb.Name != "demo" {
		t.Errorf("Name want demo, got %q", vb.Name)
	}
	if vb.Port != 8080 {
		t.Errorf("Port want 8080, got %d", vb.Port)
	}
	if vb.Rate != 0.5 {
		t.Errorf("Rate want 0.5, got %v", vb.Rate)
	}
	if !vb.Enabled {
		t.Errorf("Enabled want true, got %v", vb.Enabled)
	}
	if vb.Duration != 30000 {
		t.Errorf("Duration want 30000, got %d", vb.Duration)
	}
}

// ===== value 标签缺省不报错 =====

type valueMissingBean struct {
	Optional string `value:"missing.key"`
}

// TestValueInjection_MissingKey 配置项缺失时不报错，字段保持零值
func TestValueInjection_MissingKey(t *testing.T) {
	c := New()
	c.Provide(valueMissingBean{})
	c.Load()
	bean, _ := c.GetBean("valueMissingBean")
	if bean.(*valueMissingBean).Optional != "" {
		t.Error("want zero value for missing key")
	}
}

// ===== Destroy 倒序触发 =====

type destroyBeanA struct{}
type destroyBeanB struct{}

func (*destroyBeanA) Destroy() { testRecorder.add("destroy-a") }
func (*destroyBeanB) Destroy() { testRecorder.add("destroy-b") }

// TestLifecycle_DestroyReverseOrder 销毁按注册倒序
func TestLifecycle_DestroyReverseOrder(t *testing.T) {
	testRecorder.reset()
	c := New()
	c.Provide(destroyBeanA{})
	c.Provide(destroyBeanB{})
	c.Load()
	// 模拟销毁（Serve 会阻塞，直接调内部 destroyBeans）
	c.destroyBeans()

	want := []string{"destroy-b", "destroy-a"}
	got := testRecorder.events
	if len(got) != len(want) {
		t.Fatalf("want %d events, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] want %q, got %q (all: %v)", i, want[i], got[i], got)
		}
	}
}
