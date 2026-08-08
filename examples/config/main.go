// 示例：配置注入（value 标签 + van 配置管理）
//
// 演示通过 value 标签注入配置项，以及类型自动转换。
package main

import (
	"fmt"
	"time"

	"github.com/cheivin/di"
)

// AppConfig 使用 value 标签声明配置项
type AppConfig struct {
	Port      int           `value:"app.port"`
	Host      string        `value:"app.host"`
	Debug     bool          `value:"app.debug"`
	Timeout   time.Duration `value:"app.timeout"`   // "5s" 直接解析
	Retries   int           `value:"app.retries"`   // 字符串转 int
	Rate      float64       `value:"app.rate"`      // 字符串转 float64
	KeepAlive time.Duration `value:"app.keepalive"` // 纯数字按毫秒兜底
}

func main() {
	// 设置配置项（支持 map 批量设置、层级 key）
	di.SetDefaultPropertyMap(map[string]any{
		"app.port": 8080,
		"app.host": "localhost",
	})
	// SetProperty 覆盖 SetDefaultProperty
	di.SetPropertyMap(map[string]any{
		"app.debug":   "true", // 字符串转 bool
		"app.timeout": "5s",   // Duration 带单位
		"app.retries": "3",    // 字符串转 int
		"app.rate":    "0.5",  // 字符串转 float64
		"app.keepalive": 30000, // 纯数字 → 30000ms（毫秒兜底）
	})

	di.Provide(AppConfig{})
	di.Load()

	cfg, _ := di.GetBean("appConfig")
	c := cfg.(*AppConfig)

	fmt.Printf("Host:     %q\n", c.Host)
	fmt.Printf("Port:     %d\n", c.Port)
	fmt.Printf("Debug:    %v\n", c.Debug)
	fmt.Printf("Timeout:  %v\n", c.Timeout)
	fmt.Printf("Retries:  %d\n", c.Retries)
	fmt.Printf("Rate:     %v\n", c.Rate)
	fmt.Printf("KeepAlive:%v (30000ms 兜底)\n", c.KeepAlive)

	// 也可以直接用 LoadProperties 把配置加载到独立结构体（不注册为 bean）
	fmt.Println("\n=== LoadProperties 直接加载 ===")
	type DBConfig struct {
		Host string `value:"host"`
		Port int    `value:"port"`
	}
	di.SetDefaultProperty("mysql", map[string]any{
		"host": "127.0.0.1",
		"port": 3306,
	})
	result := di.LoadProperties("mysql.", &DBConfig{}).(DBConfig)
	fmt.Printf("DBConfig: %s:%d\n", result.Host, result.Port)
}
