package di

import (
	"errors"
	"reflect"
	"testing"
)

// 测试用的 Primary bean
type primaryBean struct{ name string }
type normalBean struct{ name string }

func (p *primaryBean) IsPrimary() bool { return true }
func (n *normalBean) IsPrimary() bool  { return false }

func buildCandidates() []BeanWithName {
	return []BeanWithName{
		{Name: "first", Bean: &normalBean{"first"}},
		{Name: "second", Bean: &primaryBean{"second"}}, // Primary
		{Name: "third", Bean: &normalBean{"third"}},
	}
}

func TestSelector_LastRegistered(t *testing.T) {
	c := buildCandidates()
	idx, err := LastRegistered{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		t.Fatal(err)
	}
	if idx != 2 {
		t.Fatalf("want idx 2, got %d", idx)
	}
}

func TestSelector_FirstRegistered(t *testing.T) {
	c := buildCandidates()
	idx, err := FirstRegistered{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("want idx 0, got %d", idx)
	}
}

func TestSelector_PrimaryFirst_HitPrimary(t *testing.T) {
	c := buildCandidates()
	idx, err := PrimaryFirst{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatalf("want idx 1 (primary), got %d", idx)
	}
}

func TestSelector_PrimaryFirst_NoPrimaryFallback(t *testing.T) {
	c := []BeanWithName{
		{Name: "a", Bean: &normalBean{"a"}},
		{Name: "b", Bean: &normalBean{"b"}},
	}
	idx, err := PrimaryFirst{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		t.Fatal(err)
	}
	// 无 Primary 时回退到取最后一个（兼容默认）
	if idx != 1 {
		t.Fatalf("want idx 1 (fallback last), got %d", idx)
	}
}

func TestSelector_PrimaryFirst_MultiplePrimaryError(t *testing.T) {
	c := []BeanWithName{
		{Name: "a", Bean: &primaryBean{"a"}},
		{Name: "b", Bean: &primaryBean{"b"}},
	}
	_, err := PrimaryFirst{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err == nil {
		t.Fatal("want error for multiple primary, got nil")
	}
	if !errors.Is(err, ErrBean) {
		t.Fatalf("want ErrBean, got %v", err)
	}
}

func TestSelector_ErrorOnAmbiguous_OneCandidate(t *testing.T) {
	c := []BeanWithName{{Name: "only", Bean: &normalBean{"only"}}}
	idx, err := ErrorOnAmbiguous{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatalf("want idx 0, got %d", idx)
	}
}

func TestSelector_ErrorOnAmbiguous_MultipleError(t *testing.T) {
	c := buildCandidates()
	_, err := ErrorOnAmbiguous{}.Select(c, reflect.TypeOf((*any)(nil)).Elem())
	if err == nil {
		t.Fatal("want error for ambiguous, got nil")
	}
	if !errors.Is(err, ErrBean) {
		t.Fatalf("want ErrBean, got %v", err)
	}
}

// 集成测试：PrimaryFirst 策略下容器实际选中 primary bean
type animal interface {
	Sound() string
}

type dog struct{}
type cat struct{}

func (dog) Sound() string { return "woof" }
func (cat) Sound() string { return "meow" }
func (cat) IsPrimary() bool { return true } // cat 是首选

type zoo struct {
	Pet animal `aware:""`
}

func TestSelector_Integration_PrimaryFirst(t *testing.T) {
	c := New()
	c.WithBeanSelector(PrimaryFirst{})
	c.Provide(dog{})
	c.Provide(cat{})
	c.Provide(zoo{})
	c.Load()

	bean, ok := c.GetBean("zoo")
	if !ok {
		t.Fatal("expected zoo bean")
	}
	z := bean.(*zoo)
	if z.Pet.Sound() != "meow" {
		t.Fatalf("want primary cat sound 'meow', got %q", z.Pet.Sound())
	}
}
