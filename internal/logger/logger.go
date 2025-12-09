package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel 定义日志级别类型
type LogLevel int

const (
	// 定义日志级别常量
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String 返回日志级别对应的字符串
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志结构体
type Logger struct {
	level      LogLevel
	output     *log.Logger
	outputFile *os.File
}

// NewLogger 创建新的日志实例
func NewLogger(levelStr string) *Logger {
	// 解析日志级别
	level := parseLogLevel(levelStr)

	var logOutput *log.Logger
	var outputFile *os.File

	// 默认输出到标准输出
	logOutput = log.New(os.Stdout, "", 0)

	return &Logger{
		level:      level,
		output:     logOutput,
		outputFile: outputFile,
	}
}

// parseLogLevel 将字符串解析为日志级别
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// SetOutputFile 设置日志输出到文件
func (l *Logger) SetOutputFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("无法打开日志文件: %v", err)
	}

	// 如果已经有打开的文件，先关闭它
	if l.outputFile != nil {
		l.outputFile.Close()
	}

	l.outputFile = file
	writer := io.MultiWriter(os.Stdout, file)
	l.output = log.New(writer, "", 0)

	return nil
}

// Close 关闭日志系统（关闭打开的文件）
func (l *Logger) Close() error {
	if l.outputFile != nil {
		return l.outputFile.Close()
	}
	return nil
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	l.output.Printf("[%s] [%s] %s\n", timestamp, level.String(), message)
}

// Debug 记录调试信息
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 记录一般信息
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 记录警告信息
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 记录错误信息
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// shortenText 缩短文本以便在日志中显示
func (l *Logger) ShortenText(text string) string {
	if len(text) <= 50 {
		return text
	}
	return text[:20] + "..." + text[len(text)-20:]
}

// GetLevel 返回当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}
