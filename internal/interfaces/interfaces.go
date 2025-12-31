package interfaces

import (
	"github.com/lyj404/win-path-convert/internal/logger"
)

// IPathConverter 路径转换器接口
// 定义了路径转换所需的基本操作
type IPathConverter interface {
	// ShouldConvert 判断是否应该转换给定的文本
	ShouldConvert(text string) bool

	// Convert 将Windows路径转换为Unix风格路径
	Convert(text string) string

	// UpdateExcludePatterns 更新排除模式
	UpdateExcludePatterns(patterns []string)
}

// IClipboardManager 剪贴板管理器接口
// 定义了剪贴板操作的基本方法
type IClipboardManager interface {
	// GetText 获取剪贴板文本内容
	GetText() (string, error)

	// SetText 设置剪贴板文本内容
	SetText(text string) error

	// HasChanged 检查剪贴板内容是否已变化
	HasChanged() (bool, error)

	// LastContentHash 返回最近一次内容的哈希
	LastContentHash() string

	// SetLastContentHash 设置最近一次内容的哈希
	SetLastContentHash(hash string)
}

// ILogger 日志接口
// 定义了日志记录的基本方法
type ILogger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	ShortenText(text string) string
}

// 确保具体的 Logger 实现满足接口
var _ ILogger = (*logger.Logger)(nil)
