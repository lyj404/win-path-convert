package config

import "win-path-convert/internal/logger"

// GlobalLogger 全局日志实例
var GlobalLogger = logger.NewLogger("info")

// InitLogger 初始化全局日志
func InitLogger(level string) {
	GlobalLogger = logger.NewLogger(level)
}

// SetLogFile 设置日志输出文件
func SetLogFile(filePath string) error {
	return GlobalLogger.SetOutputFile(filePath)
}

// CloseLogger 关闭日志系统
func CloseLogger() error {
	return GlobalLogger.Close()
}

// Log 根据级别输出日志
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
		GlobalLogger.Info(format, args...)
	}
}

func DebugLog(format string, args ...interface{}) {
	GlobalLogger.Debug(format, args...)
}

func InfoLog(format string, args ...interface{}) {
	GlobalLogger.Info(format, args...)
}

func WarnLog(format string, args ...interface{}) {
	GlobalLogger.Warn(format, args...)
}

func ErrorLog(format string, args ...interface{}) {
	GlobalLogger.Error(format, args...)
}

// ShortenText 缩短文本以便在日志中显示
func ShortenText(text string) string {
	return GlobalLogger.ShortenText(text)
}
