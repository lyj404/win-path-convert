package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// 测试轮询间隔
	if cfg.PollInterval != 100*time.Millisecond {
		t.Errorf("expected PollInterval to be 100ms, got %v", cfg.PollInterval)
	}

	// 测试自动转换开关
	if !cfg.AutoConvert {
		t.Error("expected AutoConvert to be true")
	}

	// 测试显示通知开关
	if !cfg.ShowNotifications {
		t.Error("expected ShowNotifications to be true")
	}

	// 测试日志级别
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	// 测试排除模式不为空
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected ExcludePatterns to have at least one pattern")
	}

	// 测试互斥量名称
	if cfg.MutexName == "" {
		t.Error("expected MutexName to be set")
	}
}

func TestDefaultConfig_ExcludePatterns(t *testing.T) {
	cfg := DefaultConfig()

	// 检查是否有 http:// 和 https:// 排除模式
	hasHTTP := false
	hasHTTPS := false
	for _, pattern := range cfg.ExcludePatterns {
		if pattern == "http://*" {
			hasHTTP = true
		}
		if pattern == "https://*" {
			hasHTTPS = true
		}
	}

	if !hasHTTP {
		t.Error("expected ExcludePatterns to contain 'http://*'")
	}
	if !hasHTTPS {
		t.Error("expected ExcludePatterns to contain 'https://*'")
	}
}
