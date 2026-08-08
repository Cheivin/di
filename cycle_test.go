package di

import (
	"errors"
	"strings"
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

// ===== 直接环 A <-> B =====
// aware 留空：容器按类型推断 beanName（GetBeanName(*B) == "cycleB"），与 definitionMap 对齐

type cycleA struct {
	B *cycleB `aware:""`
}
type cycleB struct {
	A *cycleA `aware:""`
}

func TestCycle_DirectCircle(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleA{})
	c.Provide(cycleB{})

	defer func() {
		err := recover().(error)
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("want ErrCircularDependency, got %v", err)
		}
		msg := err.Error()
		if !strings.Contains(msg, "cycleA") || !strings.Contains(msg, "cycleB") {
			t.Fatalf("cycle chain should mention cycleA and cycleB, got %s", msg)
		}
		t.Logf("detected cycle: %s", msg)
	}()
	c.Load()
}

// ===== 间接环 A -> B -> C -> A =====

type cycleInA struct {
	B *cycleInB `aware:""`
}
type cycleInB struct {
	C *cycleInC `aware:""`
}
type cycleInC struct {
	A *cycleInA `aware:""`
}

func TestCycle_IndirectCircle(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleInA{})
	c.Provide(cycleInB{})
	c.Provide(cycleInC{})

	defer func() {
		err := recover().(error)
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("want ErrCircularDependency, got %v", err)
		}
		t.Logf("detected cycle: %s", err.Error())
	}()
	c.Load()
}

// ===== 自依赖 A -> A =====

type cycleSelf struct {
	Self *cycleSelf `aware:""`
}

func TestCycle_SelfReference(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleSelf{})

	defer func() {
		err := recover().(error)
		if !errors.Is(err, ErrCircularDependency) {
			t.Fatalf("want ErrCircularDependency, got %v", err)
		}
		t.Logf("detected cycle: %s", err.Error())
	}()
	c.Load()
}

// ===== 无环：正常加载 =====

type cycleDB struct{}

type cycleUserDao struct {
	Db *cycleDB `aware:""`
}

func TestCycle_NoCycle(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(cycleDB{})
	c.Provide(cycleUserDao{})
	c.Load() // 不应 panic
	if _, ok := c.GetBean("cycleUserDao"); !ok {
		t.Fatal("expected cycleUserDao bean to be loaded")
	}
}

// ===== 两条独立链，无环 =====

type chainA struct {
	B *chainB `aware:""`
}
type chainB struct{}
type chainC struct {
	D *chainD `aware:""`
}
type chainD struct{}

func TestCycle_IndependentChains(t *testing.T) {
	c := newCycleTestContainer(t)
	c.Provide(chainA{})
	c.Provide(chainB{})
	c.Provide(chainC{})
	c.Provide(chainD{})
	c.Load() // 不应 panic
}
