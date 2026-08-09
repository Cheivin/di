package di

import (
	"errors"
	"testing"
)

// cycleTestLogger 收集日志并把 Fatal 转为 panic，便于断言
type cycleTestLogger struct {
	t *testing.T
}

func (l cycleTestLogger) DebugMode(bool) {}
func (l cycleTestLogger) Debug(s string) { l.t.Log("[debug]", s) }
func (l cycleTestLogger) Info(s string)  { l.t.Log("[info]", s) }
func (l cycleTestLogger) Warn(s string)  { l.t.Log("[warn]", s) }
func (l cycleTestLogger) Fatal(err error) {
	l.t.Log("[fatal]", err)
	panic(err)
}

func newCycleTestContainer(t *testing.T) *di {
	c := New()
	c.Log(cycleTestLogger{t: t})
	return c
}

// ===== 默认行为：指针循环依赖正常注入（di 两阶段设计天然支持）=====

type cycleA struct {
	B *cycleB `aware:""`
}
type cycleB struct {
	A *cycleA `aware:""`
}

// TestCycle_PtrMutual_Default 默认（不检测）时 A↔B 正常注入
func TestCycle_PtrMutual_Default(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleA{})
	c.Provide(cycleB{})
	c.Load() // 不应 panic

	a, _ := c.GetBean("cycleA")
	if a == nil {
		t.Fatal("expected cycleA bean")
	}
	ra := a.(*cycleA)
	if ra.B == nil {
		t.Fatal("expected A.B injected")
	}
	if ra.B.A != ra {
		t.Fatal("expected A.B.A == A (circular reference)")
	}
}

// 间接环 A → B → C → A
type cycleInA struct {
	B *cycleInB `aware:""`
}
type cycleInB struct {
	C *cycleInC `aware:""`
}
type cycleInC struct {
	A *cycleInA `aware:""`
}

func TestCycle_Indirect_Default(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleInA{})
	c.Provide(cycleInB{})
	c.Provide(cycleInC{})
	c.Load()

	a, _ := c.GetBean("cycleInA")
	ra := a.(*cycleInA)
	if ra.B == nil || ra.B.C == nil || ra.B.C.A != ra {
		t.Fatal("expected circular reference chain A.B.C.A == A")
	}
}

// 自依赖 A → A
type cycleSelf struct {
	Self *cycleSelf `aware:""`
}

func TestCycle_SelfReference_Default(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleSelf{})
	c.Load()

	a, _ := c.GetBean("cycleSelf")
	ra := a.(*cycleSelf)
	if ra.Self != ra {
		t.Fatal("expected Self == self (circular reference)")
	}
}

// ===== opt-in：WithCircularCheck(true) 时检测并报错 =====

func TestCycle_DetectDirectCircle_WhenEnabled(t *testing.T) {
	c := newCycleTestContainer(t)
	c.WithCircularCheck(true)
	c.Provide(cycleA{})
	c.Provide(cycleB{})

	defer func() {
		err := recover().(error)
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("want ErrCircularDependency, got %v", err)
		}
		t.Logf("检测到循环: %s", err.Error())
	}()
	c.Load()
}

func TestCycle_SelfReference_WhenEnabled(t *testing.T) {
	c := newCycleTestContainer(t)
	c.WithCircularCheck(true)
	c.Provide(cycleSelf{})

	defer func() {
		err := recover().(error)
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("want ErrCircularDependency, got %v", err)
		}
		t.Logf("检测到循环: %s", err.Error())
	}()
	c.Load()
}

// ===== 无环：检测开启也能正常加载 =====

type cycleDB struct{}
type cycleUserDao struct {
	Db *cycleDB `aware:""`
}

func TestCycle_NoCycle(t *testing.T) {
	c := newCycleTestContainer(t)
	c.WithCircularCheck(true) // 开启检测
	c.Provide(cycleDB{})
	c.Provide(cycleUserDao{})
	c.Load() // 无环，不应 panic

	if _, ok := c.GetBean("cycleUserDao"); !ok {
		t.Fatal("expected cycleUserDao bean")
	}
}
