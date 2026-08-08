package di

import (
	"os"
	"strings"
	"testing"
)

// TestAutoMigrateEnv 环境变量自动注入配置（接口方法）
func TestAutoMigrateEnv(t *testing.T) {
	os.Setenv("DI_TEST_HOST", "localhost")
	os.Setenv("DI_TEST_PORT", "8080")
	os.Setenv("DI_TEST_DEBUG", "true")
	defer func() {
		os.Unsetenv("DI_TEST_HOST")
		os.Unsetenv("DI_TEST_PORT")
		os.Unsetenv("DI_TEST_DEBUG")
	}()

	c := New()
	c.AutoMigrateEnv()

	// 下划线应转为点号
	if v := c.GetProperty("di.test.host"); v != "localhost" {
		t.Errorf("host want localhost, got %v", v)
	}
	if v := c.GetProperty("di.test.port"); v != "8080" {
		t.Errorf("port want 8080, got %v", v)
	}
	if v := c.GetProperty("di.test.debug"); v != "true" {
		t.Errorf("debug want true, got %v", v)
	}
}

// TestLoadEnvironment_WithPrefix prefix 过滤 + 去前缀 + 替换
func TestLoadEnvironment_WithPrefix(t *testing.T) {
	os.Setenv("MYAPP_DB_HOST", "db.local")
	os.Setenv("OTHER_KEY", "ignored")
	defer func() {
		os.Unsetenv("MYAPP_DB_HOST")
		os.Unsetenv("OTHER_KEY")
	}()

	// 只读 MYAPP_ 前缀，trimPrefix 去掉前缀，_ 转 .
	// 注意：LoadEnvironment 不做大小写折叠（那是 van 的职责），key 保留原大小写
	envMap := LoadEnvironment(
		strings.NewReplacer("_", "."),
		true,
		"MYAPP_",
	)

	// trimPrefix MYAPP_ 后剩 DB_HOST，replacer 得 DB.HOST（大写保留）
	if v, ok := envMap["DB.HOST"]; !ok || v != "db.local" {
		t.Errorf("DB.HOST want db.local, got %v (ok=%v)", v, ok)
	}
	if _, ok := envMap["OTHER.KEY"]; ok {
		t.Error("OTHER_KEY should be filtered out")
	}
}

// TestLoadEnvironment_NoReplacer replacer 为 nil 时保留原始 key
func TestLoadEnvironment_NoReplacer(t *testing.T) {
	os.Setenv("PLAIN_KEY", "value")
	defer os.Unsetenv("PLAIN_KEY")

	envMap := LoadEnvironment(nil, false, "PLAIN_")
	if v, ok := envMap["PLAIN_KEY"]; !ok || v != "value" {
		t.Errorf("PLAIN_KEY want value, got %v (ok=%v)", v, ok)
	}
}
