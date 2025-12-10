package config

import "github.com/lyj404/win-path-convert/internal/logger"

// GlobalLogger 全局日志实例
// 这是整个应用程序使用的唯一日志记录器实例，所有组件都通过这个实例输出日志
// 初始设置为info级别，可以根据需要通过InitLogger函数调整
var GlobalLogger *logger.Logger = logger.NewLogger("info")

// InitLogger 初始化全局日志系统
// 该函数根据指定的日志级别创建新的全局日志记录器，替换默认实例
// 在应用程序启动时应调用该函数设置合适的日志级别
// 参数:
//   - level: 日志级别，可以是 "debug", "info", "warn", "error" 之一
func InitLogger(level string) {
	// 创建指定级别的新日志记录器并替换当前的全局日志实例
	GlobalLogger = logger.NewLogger(level)
}

// SetLogFile 设置日志输出文件
// 该函数将日志输出从控制台重定向到指定的文件，便于持久化存储和后续分析
// 日志会同时输出到控制台和文件，确保用户既能看到日志内容，又能保存到文件
// 参数:
//   - filePath: 日志文件的完整路径，可以是绝对路径或相对路径
//
// 返回值:
//   - error: 如果设置文件输出失败，返回相应的错误信息
func SetLogFile(filePath string) error {
	return GlobalLogger.SetOutputFile(filePath)
}

// CloseLogger 关闭日志系统
// 该函数关闭全局日志记录器，确保所有缓冲的日志信息被写入存储设备
// 在应用程序退出前应调用此函数，防止日志信息丢失
// 返回值:
//   - error: 如果关闭过程中发生错误，返回相应的错误信息
func CloseLogger() error {
	return GlobalLogger.Close()
}

// Log 根据级别输出日志
// 这是一个通用的日志函数，根据指定的级别输出格式化日志信息
// 当不确定日志级别或需要动态决定日志级别时非常有用
// 注意：这个函数主要作为便捷包装器存在，实际使用中推荐使用 GlobalLogger 实例的方法
// 参数:
//   - level: 日志级别，可以是 "debug", "info", "warn", "error" 之一
//   - format: 格式化字符串，类似于 fmt.Printf 的格式
//   - args: 格式化参数，用于填充格式化字符串中的占位符
func Log(level string, format string, args ...interface{}) {
	switch level {
	case "debug":
		GlobalLogger.Debug(format, args...)
	case "info":
		GlobalLogger.Info(format, args...)
	case "warn":
		GlobalLogger.Warn(format, args...)
	case "error":
		GlobalLogger.Error(format, args...)
	default:
		// 如果提供了未知的日志级别，默认使用info级别
		GlobalLogger.Info(format, args...)
	}
}

// ShortenText 缩短文本以便在日志中显示
// 该函数将过长的文本截断为适当长度，防止日志输出变得过于冗长
// 特别适用于显示路径或URL等可能很长的文本内容
// 注意：这个函数主要作为便捷包装器存在，实际使用中推荐使用 GlobalLogger.ShortenText()
// 参数:
//   - text: 需要缩短的原始文本
//
// 返回值:
//   - string: 缩短后的文本，保留了原文的关键信息
func ShortenText(text string) string {
	return GlobalLogger.ShortenText(text)
}
