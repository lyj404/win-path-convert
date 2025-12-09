package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"win-path-convert/internal/app"
	"win-path-convert/internal/clipboard"
	"win-path-convert/internal/config"
	"win-path-convert/internal/logger"
	"win-path-convert/internal/pathconv"
	"win-path-convert/internal/singleton"
	"win-path-convert/internal/winapi"
)

// PathConvertApp 聚合应用依赖与运行状态
type PathConvertApp struct {
	cfg    *config.Config
	log    *logger.Logger
	cb     *clipboard.ClipboardManager
	pc     *pathconv.PathConverter
	ctx    context.Context
	cancel context.CancelFunc
	sigCh  chan os.Signal
}

// NewPathConvertApp 创建应用实例
func NewPathConvertApp(cfg *config.Config, log *logger.Logger) *PathConvertApp {
	ctx, cancel := context.WithCancel(context.Background())
	return &PathConvertApp{
		cfg:    cfg,
		log:    log,
		cb:     clipboard.NewClipboardManager(),
		ctx:    ctx,
		cancel: cancel,
		sigCh:  make(chan os.Signal, 1),
	}
}

// Initialize 初始化组件
func (a *PathConvertApp) Initialize() error {
	a.log.Info("初始化Windows路径转换工具...")
	a.pc = pathconv.NewPathConverter(a.cfg.ExcludePatterns, a.log)
	signal.Notify(a.sigCh, syscall.SIGINT, syscall.SIGTERM)
	return nil
}

// Cleanup 释放资源
func (a *PathConvertApp) Cleanup() {
	a.log.Info("正在清理资源...")
	if a.cancel != nil {
		a.cancel()
	}
	singleton.ReleaseSingleton()
	config.CloseLogger()
}

// Run 运行应用主循环
func (a *PathConvertApp) Run() error {
	a.log.Info("应用程序已启动，按Ctrl+C退出程序")

	if runtime.GOOS != "windows" {
		return fmt.Errorf("此程序只能在Windows系统上运行")
	}

	if err := a.runWithClipboardListener(); err != nil {
		a.log.Warn("无法使用剪贴板监听API，回退到轮询模式: %v", err)
		return a.runWithPolling()
	}
	return nil
}

// runWithClipboardListener 通过隐藏窗口监听剪贴板
func (a *PathConvertApp) runWithClipboardListener() error {
	a.log.Info("使用剪贴板监听模式")
	tid := getCurrentThreadID()

	className, _ := syscall.UTF16PtrFromString("PathConvertClipboardListener")
	hInstance, _, _ := winapi.ProcGetModuleHandleW.Call(0)

	wndClass := app.WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(app.WndClassEx{})),
		LpfnWndProc:   syscall.NewCallback(a.windowProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}

	if ret, _, err := winapi.ProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass))); ret == 0 {
		return fmt.Errorf("注册窗口类失败: %v", err)
	}
	defer winapi.ProcUnregisterClassW.Call(uintptr(unsafe.Pointer(className)), hInstance)

	hwnd, _, err := winapi.ProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0,
		0,
		0,
		hInstance,
		uintptr(unsafe.Pointer(a)),
	)
	if hwnd == 0 {
		return fmt.Errorf("创建隐藏窗口失败: %v", err)
	}
	defer winapi.ProcDestroyWindow.Call(hwnd)

	if ret, _, err := winapi.ProcAddClipboardFormatListener.Call(hwnd); ret == 0 {
		return fmt.Errorf("无法注册剪贴板监听: %v", err)
	}
	defer winapi.ProcRemoveClipboardFormatListener.Call(hwnd)

	// 监听退出信号并唤醒消息循环
	go func(tid uint32) {
		select {
		case <-a.sigCh:
			postQuitToThread(tid)
		case <-a.ctx.Done():
			postQuitToThread(tid)
		}
	}(tid)

	var m app.Msg
	for {
		select {
		case <-a.ctx.Done():
			return nil
		default:
		}

		ret, _, err := winapi.ProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("消息循环错误: %v", err)
		}
		if ret == 0 {
			break
		}

		if m.Message == app.WMClipboardUpdate {
			a.processClipboardChange()
		}

		winapi.ProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		winapi.ProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	return nil
}

// windowProc 处理窗口消息
func (a *PathConvertApp) windowProc(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	switch message {
	case app.WMDestroy:
		winapi.ProcPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := winapi.ProcDefWindowProcW.Call(hwnd, uintptr(message), wparam, lparam)
	return ret
}

// runWithPolling 轮询模式
func (a *PathConvertApp) runWithPolling() error {
	a.log.Info("使用轮询模式，间隔: %v", a.cfg.PollInterval)
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			changed, err := a.cb.HasChanged()
			if err != nil {
				a.log.Debug("检查剪贴板时出错: %v", err)
				continue
			}
			if changed {
				a.processClipboardChange()
			}
		case <-a.sigCh:
			a.log.Info("收到停止信号")
			if a.cancel != nil {
				a.cancel()
			}
			return nil
		case <-a.ctx.Done():
			return nil
		}
	}
}

// processClipboardChange 处理剪贴板变化
func (a *PathConvertApp) processClipboardChange() {
	a.log.Debug("检测到剪贴板变化")
	if !a.cfg.AutoConvert {
		a.log.Debug("自动转换已禁用，忽略变化")
		return
	}

	rawText, err := a.cb.GetText()
	if err != nil {
		a.log.Debug("无法获取剪贴板内容: %v", err)
		return
	}
	normalized := strings.TrimSpace(rawText)

	currentHash := clipboard.QuickHash(normalized)
	if currentHash == a.cb.LastContentHash() {
		a.log.Debug("剪贴板内容未变化，跳过处理")
		return
	}

	if !a.pc.ShouldConvert(rawText) {
		a.log.Debug("不需要转换的内容: %s", config.ShortenText(rawText))
		a.cb.SetLastContentHash(currentHash)
		return
	}

	converted := a.pc.Convert(rawText)
	if converted != rawText {
		if err := a.cb.SetText(converted); err != nil {
			a.log.Error("无法设置剪贴板内容: %v", err)
			return
		}

		if a.cfg.ShowNotifications {
			a.log.Info("已转换路径:")
			a.log.Info("  原路径: %s", rawText)
			a.log.Info("  转换后: %s", converted)
		} else {
			a.log.Debug("已转换路径，但不显示通知")
		}

		a.cb.SetLastContentHash(clipboard.QuickHash(converted))
		return
	}

	a.cb.SetLastContentHash(currentHash)
}

func getCurrentThreadID() uint32 {
	ret, _, _ := winapi.ProcGetCurrentThreadId.Call()
	return uint32(ret)
}

func postQuitToThread(tid uint32) {
	winapi.ProcPostThreadMessage.Call(uintptr(tid), uintptr(app.WMQuit), 0, 0)
}

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("此程序只能在Windows系统上运行")
		return
	}

	cfg := config.DefaultConfig()
	singleton.SetMutexName(cfg.MutexName)
	if !singleton.InitSingleton() {
		fmt.Println("程序已在运行中")
		return
	}
	defer singleton.ReleaseSingleton()

	config.InitLogger(cfg.LogLevel)
	defer config.CloseLogger()
	appLogger := logger.NewLogger(cfg.LogLevel)

	app := NewPathConvertApp(cfg, appLogger)
	if err := app.Initialize(); err != nil {
		appLogger.Error("应用程序初始化失败: %v", err)
		return
	}
	defer app.Cleanup()

	appLogger.Info("Windows路径自动转换工具已启动")
	appLogger.Info("复制包含反斜杠的路径时，将自动转换为正斜杠格式")
	appLogger.Info("日志级别: %s", cfg.LogLevel)
	appLogger.Info("自动转换: %t", cfg.AutoConvert)
	appLogger.Info("显示通知: %t", cfg.ShowNotifications)
	appLogger.Info("按Ctrl+C或Ctrl+Break退出程序")

	if err := app.Run(); err != nil {
		appLogger.Error("运行错误: %v", err)
		return
	}

	appLogger.Info("应用程序已正常退出")
}
