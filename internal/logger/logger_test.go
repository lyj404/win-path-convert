package logger

import (
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"DEBUG level", DEBUG, "DEBUG"},
		{"INFO level", INFO, "INFO"},
		{"WARN level", WARN, "WARN"},
		{"ERROR level", ERROR, "ERROR"},
		{"Unknown level", LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		levelStr string
		expected LogLevel
	}{
		{"debug", "debug", DEBUG},
		{"DEBUG", "DEBUG", DEBUG},
		{"info", "info", INFO},
		{"INFO", "INFO", INFO},
		{"warn", "warn", WARN},
		{"warning", "warning", WARN},
		{"WARN", "WARN", WARN},
		{"error", "error", ERROR},
		{"ERROR", "ERROR", ERROR},
		{"unknown", "unknown", INFO}, // 默认返回 INFO
		{"", "", INFO},               // 空字符串默认返回 INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLogLevel(tt.levelStr); got != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.levelStr, got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger("debug")
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	if logger.level != DEBUG {
		t.Errorf("expected level DEBUG, got %v", logger.level)
	}
}

func TestShortenText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"Short text", "hello", "hello"},
		{"Exact 50 chars", "12345678901234567890123456789012345678901234567890", "12345678901234567890123456789012345678901234567890"},
		{"Long text", "1234567890123456789012345678901234567890123456789012345", "12345678901234567890...67890123456789012345"},
	}

	logger := NewLogger("info")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logger.ShortenText(tt.text); got != tt.expected {
				t.Errorf("ShortenText() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	logger := NewLogger("debug")

	// 测试所有日志方法不会 panic
	logger.Debug("debug message %d", 1)
	logger.Info("info message %s", "test")
	logger.Warn("warn message %.2f", 3.14)
	logger.Error("error message %v", map[string]int{"a": 1})
}

func TestLogger_LevelFiltering(t *testing.T) {
	logger := NewLogger("error") // 只显示 ERROR 级别

	// 这些方法不会 panic，只是不会输出
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	// 由于 logger.output 是标准输出，我们无法捕获它
	// 这里只测试方法不会 panic
}

func TestGetLevelAndSetLevel(t *testing.T) {
	logger := NewLogger("info")

	if logger.GetLevel() != INFO {
		t.Errorf("expected INFO level, got %v", logger.GetLevel())
	}

	logger.SetLevel(ERROR)

	if logger.GetLevel() != ERROR {
		t.Errorf("expected ERROR level after SetLevel, got %v", logger.GetLevel())
	}
}
