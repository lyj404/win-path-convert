package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/lyj404/win-path-convert/internal/clipboard"
	"github.com/lyj404/win-path-convert/internal/config"
	"github.com/lyj404/win-path-convert/internal/interfaces"
	"github.com/lyj404/win-path-convert/internal/logger"
	"github.com/lyj404/win-path-convert/internal/pathconv"
	"github.com/lyj404/win-path-convert/internal/singleton"
)

// PathConvertApp 聚合应用依赖与运行状态
type PathConvertApp struct {
	cfg    *config.Config               // 应用配置对象，包含用户设置的各种参数
	log    *logger.Logger               // 日志记录器，用于输出应用运行信息
	cb     interfaces.IClipboardManager // 剪贴板管理器，负责监听和操作剪贴板
	pc     interfaces.IPathConverter    // 路径转换器，负责将Windows路径转换为Unix风格路径
	ctx    context.Context              // 上下文对象，用于协程间的通知和取消
	cancel context.CancelFunc           // 取消函数，用于通知所有协程停止运行
	sigCh  chan os.Signal               // 信号通道，用于接收操作系统信号（如Ctrl+C）
}

// NewPathConvertApp 创建应用实例
// 该函数初始化应用程序的核心结构体，设置基本的运行环境
// 参数:
//   - cfg: 应用配置对象
//   - log: 日志记录器
//
// 返回值:
//   - *PathConvertApp: 初始化完成的应用程序实例
func NewPathConvertApp(cfg *config.Config, log *logger.Logger) *PathConvertApp {
	// 创建上下文和对应的取消函数，用于优雅地关闭应用程序
	ctx, cancel := context.WithCancel(context.Background())
	return &PathConvertApp{
		cfg:    cfg,
		log:    log,
		cb:     clipboard.NewClipboardManager(), // 初始化剪贴板管理器
		ctx:    ctx,
		cancel: cancel,
		sigCh:  make(chan os.Signal, 1), // 创建信号通道，缓冲大小为1，防止信号丢失
	}
}

// Initialize 初始化组件
// 该函数负责初始化应用程序运行所需的各种组件
// 执行内容:
//  1. 初始化路径转换器，加载排除模式配置
//  2. 设置信号监听，捕获SIGINT和SIGTERM信号
//
// 返回值:
//   - error: 初始化过程中可能发生的错误
func (a *PathConvertApp) Initialize() error {
	a.log.Info("初始化Windows路径转换工具...")
	// 创建路径转换器实例，传入排除模式和日志记录器
	a.pc = pathconv.NewPathConverter(a.cfg.ExcludePatterns, a.log)
	// 注册信号监听，捕获SIGINT(Ctrl+C)和SIGTERM信号
	signal.Notify(a.sigCh, syscall.SIGINT, syscall.SIGTERM)
	return nil
}

// Cleanup 释放资源
// 该函数负责在应用程序退出前释放所有资源
// 执行内容:
//  1. 取消上下文，通知所有协程停止运行
//  2. 释放单例模式资源
//  3. 关闭日志记录器
func (a *PathConvertApp) Cleanup() {
	a.log.Info("正在清理资源...")
	// 调用取消函数，通知所有监听ctx.Done()的协程退出
	if a.cancel != nil {
		a.cancel()
	}
	// 释放单例锁，允许下一个程序实例启动
	singleton.ReleaseSingleton()
	// 关闭日志记录器，确保日志信息被写入文件
	config.CloseLogger()
}

// Run 运行应用主循环
// 这是应用程序的主要运行函数，负责启动剪贴板监听服务
// 执行内容:
//  1. 首先检查当前操作系统是否为Windows
//  2. 尝试使用剪贴板监听API
//  3. 如果监听API不可用，回退到轮询模式
//
// 返回值:
//   - error: 运行过程中可能发生的错误
func (a *PathConvertApp) Run() error {
	a.log.Info("应用程序已启动，按Ctrl+C退出程序")

	// 平台检查，确保程序只在Windows系统上运行
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此程序只能在Windows系统上运行")
	}

	// 优先尝试使用Windows剪贴板监听API
	if err := a.runWithClipboardListener(); err != nil {
		a.log.Warn("无法使用剪贴板监听API，回退到轮询模式: %v", err)
		// 如果监听API不可用（例如权限不足或系统版本不支持），回退到轮询模式
		return a.runWithPolling()
	}
	return nil
}

// RunApplication 应用程序启动入口
// 这是整个应用程序的入口点，负责初始化所有组件并启动应用程序
// 执行流程:
//  1. 平台检查
//  2. 单例模式初始化
//  3. 日志系统初始化
//  4. 应用程序实例创建和初始化
//  5. 启动应用程序主循环
//  6. 清理资源
//
// 返回值:
//   - error: 运行过程中可能发生的错误
func RunApplication() error {
	// 平台检查，确保程序只在Windows系统上运行
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此程序只能在Windows系统上运行")
	}

	// 加载默认配置
	cfg := config.DefaultConfig()
	// 设置单例模式的互斥锁名称（防止多个实例同时运行）
	singleton.SetMutexName(cfg.MutexName)
	// 尝试初始化单例（获取全局锁）
	if !singleton.InitSingleton() {
		return fmt.Errorf("程序已在运行中")
	}
	// 确保退出时释放单例锁
	defer singleton.ReleaseSingleton()

	// 初始化日志系统
	config.InitLogger(cfg.LogLevel)
	// 确保退出时关闭日志系统
	defer config.CloseLogger()
	// 使用全局日志实例
	appLogger := config.GlobalLogger

	// 创建应用程序实例
	app := NewPathConvertApp(cfg, appLogger)
	// 初始化应用程序组件
	if err := app.Initialize(); err != nil {
		appLogger.Error("应用程序初始化失败: %v", err)
		return err
	}
	// 确保退出时清理资源
	defer app.Cleanup()

	// 输出应用程序启动信息
	appLogger.Info("Windows路径自动转换工具已启动")
	appLogger.Info("复制包含反斜杠的路径时，将自动转换为正斜杠格式")
	appLogger.Info("日志级别: %s", cfg.LogLevel)
	appLogger.Info("自动转换: %t", cfg.AutoConvert)
	appLogger.Info("显示通知: %t", cfg.ShowNotifications)
	appLogger.Info("按Ctrl+C或Ctrl+Break退出程序")

	// 运行应用程序主循环
	if err := app.Run(); err != nil {
		appLogger.Error("运行错误: %v", err)
		return err
	}

	// 应用程序正常退出
	appLogger.Info("应用程序已正常退出")
	return nil
}
