package config

import "time"

// Config 包含应用程序的所有配置选项
type Config struct {
	PollInterval      time.Duration // 轮询间隔（仅在无法使用剪贴板监听API时）
	AutoConvert       bool          // 是否自动转换路径
	ShowNotifications bool          // 是否显示转换通知
	ExcludePatterns   []string      // 排除的模式
	LogLevel          string        // 日志级别: debug, info, warn, error
	MutexName         string        // 互斥量名称，避免与其他实例冲突
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		PollInterval:      100 * time.Millisecond,
		AutoConvert:       true,
		ShowNotifications: true,
		ExcludePatterns: []string{
			"http://*", "https://*",
		},
		LogLevel:  "info",
		MutexName: "PathConvertToolMutex",
	}
}
