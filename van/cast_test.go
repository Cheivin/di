package van

import (
	"reflect"
	"testing"
	"time"
)

func Test_toString(t *testing.T) {
	t.Log(toString("hello"))
	t.Log(toString(reflect.ValueOf("hello")))
}

func Test_toStringMap(t *testing.T) {
	t.Log(toStringMap("xxxx"))
	t.Log(toStringMap(map[string]string{"1": "x", "2": "c"}))
	t.Log(toStringMap(map[string]any{"1": "x", "2": "c"}))
	t.Log(toStringMap(map[any]any{"1": "x", "2": "c"}))
	t.Log(toStringMap(map[int]any{1: "x", 2: "c"}))
}

func Test_cast(t *testing.T) {
	//t.Log(Cast("aa", reflect.TypeOf("")))
	//t.Log(Cast(65, reflect.TypeOf("")))
	//t.Log(Cast("1", reflect.TypeOf(float32(1))))
	//t.Log(Cast(65, reflect.TypeOf(float32(1))))
	//t.Log("type",reflect.TypeOf(time.Hour))
	//t.Log(Cast(65,reflect.TypeOf(time.Nanosecond)))
	t.Log(Cast("65", reflect.TypeFor[int64]()))
	t.Log(Cast("65h", reflect.TypeFor[time.Duration]()))
	t.Log(Cast("65", reflect.TypeFor[time.Duration]()))
}

// Stringer 类型
type strVal struct{ s string }

func (v strVal) String() string { return v.s }

func TestCast_Stringer(t *testing.T) {
	got, err := Cast(strVal{"42"}, reflect.TypeFor[int]())
	if err != nil {
		t.Fatal(err)
	}
	if got.(int) != 42 {
		t.Fatalf("want 42, got %v", got)
	}
	// Stringer → string
	s, err := Cast(strVal{"hello"}, reflect.TypeFor[string]())
	if err != nil {
		t.Fatal(err)
	}
	if s.(string) != "hello" {
		t.Fatalf("want hello, got %v", s)
	}
}

func TestCast_Slice(t *testing.T) {
	// []int → []string
	src := []int{1, 2, 3}
	got, err := Cast(src, reflect.TypeFor[[]string]())
	if err != nil {
		t.Fatal(err)
	}
	ss := got.([]string)
	if len(ss) != 3 || ss[0] != "1" || ss[2] != "3" {
		t.Fatalf("want [1 2 3], got %v", ss)
	}
	// string → []byte
	b, err := Cast("abc", reflect.TypeFor[[]byte]())
	if err != nil {
		t.Fatal(err)
	}
	if string(b.([]byte)) != "abc" {
		t.Fatalf("want abc, got %v", b)
	}
	// []string → []int（逐个转换）
	got2, err := Cast([]string{"10", "20"}, reflect.TypeFor[[]int]())
	if err != nil {
		t.Fatal(err)
	}
	ii := got2.([]int)
	if len(ii) != 2 || ii[0] != 10 || ii[1] != 20 {
		t.Fatalf("want [10 20], got %v", ii)
	}
}

func TestCast_Duration_MillisFallback(t *testing.T) {
	// 纯数字按毫秒兜底（历史行为）
	d, err := Cast("65", reflect.TypeFor[time.Duration]())
	if err != nil {
		t.Fatal(err)
	}
	if d.(time.Duration) != 65*time.Millisecond {
		t.Fatalf("want 65ms, got %v", d)
	}
	// 带单位直接解析
	d2, err := Cast("65h", reflect.TypeFor[time.Duration]())
	if err != nil {
		t.Fatal(err)
	}
	if d2.(time.Duration) != 65*time.Hour {
		t.Fatalf("want 65h, got %v", d2)
	}
}

func TestCast_UnknownFallback(t *testing.T) {
	// 未知类型不再静默丢值，toString 兜底 fmt.Sprint
	s := toString([]byte{1, 2, 3})
	if s == "" {
		t.Fatal("want non-empty string for []byte")
	}
	// 非字符串目标类型返回 error 而非 panic
	_, err := Cast(struct{}{}, reflect.TypeFor[string]())
	if err != nil {
		t.Fatal(err)
	}
}
