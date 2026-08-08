package di

import (
	"testing"
	"time"
)

// LoadProperties 的配置加载测试（从 di/test 迁移）
type testConfig struct {
	Username    string        `value:"username"`
	Password    string        `value:"password"`
	Host        string        `value:"host"`
	Port        int           `value:"port"`
	Database    string        `value:"database"`
	Parameters  string        `value:"parameters"`
	MaxIdle     int           `value:"pool.max-idle"`
	MaxOpen     int           `value:"pool.max-open"`
	MaxLifeTime time.Duration `value:"pool.max-life-time"`
	MaxIdleTime time.Duration `value:"pool.max-idle-time"`
}

// TestLoadProperties 配置项通过 LoadProperties 映射到结构体并做类型转换
func TestLoadProperties(t *testing.T) {
	c := New()
	c.SetDefaultProperty("gorm", map[string]any{
		"username": "root",
		"password": "root",
		"host":     "localhost",
		"port":     3306,
		"pool": map[string]any{
			"max-idle":      0,
			"max-open":      0,
			"max-life-time": "30s",
			"max-idle-time": 10000,
		},
		"log.level": 4,
	})

	cfg := testConfig{}
	// LoadProperties 返回新构造并注入完成的实例，不会回填传入的 cfg
	result := c.LoadProperties("gorm.", &cfg)
	loaded, ok := result.(testConfig)
	if !ok {
		t.Fatalf("expected testConfig, got %T", result)
	}
	if loaded.Username != "root" {
		t.Errorf("username want root, got %q", loaded.Username)
	}
	if loaded.Port != 3306 {
		t.Errorf("port want 3306, got %d", loaded.Port)
	}
	// "30s" 直接解析为 Duration
	if loaded.MaxLifeTime != 30*time.Second {
		t.Errorf("max-life-time want 30s, got %v", loaded.MaxLifeTime)
	}
	// 纯数字 10000 按毫秒兜底（历史兼容行为）
	if loaded.MaxIdleTime != 10000*time.Millisecond {
		t.Errorf("max-idle-time want 10s (10000ms), got %v", loaded.MaxIdleTime)
	}
	// 点号分隔的 key（log.level）不影响 gorm. 前缀的字段加载
}

// TestLoadProperties_Prefix 点号前缀隔离
func TestLoadProperties_Prefix(t *testing.T) {
	c := New()
	c.SetProperty("a.x", "ax")
	c.SetProperty("b.x", "bx")
	type S struct {
		X string `value:"x"`
	}
	s := c.LoadProperties("b.", &S{}).(S)
	if s.X != "bx" {
		t.Fatalf("want bx, got %q", s.X)
	}
}
