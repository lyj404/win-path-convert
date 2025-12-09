package app

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

	"github.com/lyj404/win-path-convert/internal/clipboard"
	"github.com/lyj404/win-path-convert/internal/config"
	"github.com/lyj404/win-path-convert/internal/logger"
	"github.com/lyj404/win-path-convert/internal/pathconv"
	"github.com/lyj404/win-path-convert/internal/singleton"
	"github.com/lyj404/win-path-convert/internal/winapi"
)

// PathConvertApp 聚合应用依赖与运行状态
type PathConvertApp struct {
	cfg    *config.Config              // 应用配置对象，包含用户设置的各种参数
	log    *logger.Logger              // 日志记录器，用于输出应用运行信息
	cb     *clipboard.ClipboardManager // 剪贴板管理器，负责监听和操作剪贴板
	pc     *pathconv.PathConverter     // 路径转换器，负责将Windows路径转换为Unix风格路径
	ctx    context.Context             // 上下文对象，用于协程间的通知和取消
	cancel context.CancelFunc          // 取消函数，用于通知所有协程停止运行
	sigCh  chan os.Signal              // 信号通道，用于接收操作系统信号（如Ctrl+C）
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

// runWithClipboardListener 通过隐藏窗口监听剪贴板
// 这是Windows系统下的高效实现，通过注册隐藏窗口监听剪贴板变化事件
// 工作原理:
//  1. 创建一个隐藏窗口
//  2. 注册剪贴板格式监听器
//  3. 进入消息循环，等待剪贴板变化事件
//
// 返回值:
//   - error: 初始化或运行过程中可能发生的错误
func (a *PathConvertApp) runWithClipboardListener() error {
	a.log.Info("使用剪贴板监听模式")
	// 获取当前线程ID，用于后面向特定线程发送退出消息
	tid := getCurrentThreadID()

	// 创建窗口类名称字符串（UTF-16编码，Windows API要求）
	className, _ := syscall.UTF16PtrFromString("PathConvertClipboardListener")
	// 获取当前应用程序实例句柄，用于注册窗口类
	hInstance, _, _ := winapi.ProcGetModuleHandleW.Call(0)

	// 设置窗口类结构体，定义窗口的基本属性和行为
	wndClass := WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(WndClassEx{})), // 结构体大小
		LpfnWndProc:   syscall.NewCallback(a.windowProc),   // 窗口过程函数指针
		HInstance:     hInstance,                           // 应用程序实例句柄
		LpszClassName: className,                           // 窗口类名称
	}

	// 注册窗口类，创建窗口前必须先注册窗口类
	if ret, _, err := winapi.ProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass))); ret == 0 {
		return fmt.Errorf("注册窗口类失败: %v", err)
	}
	// 确保退出时注销窗口类，释放系统资源
	defer winapi.ProcUnregisterClassW.Call(uintptr(unsafe.Pointer(className)), hInstance)

	// 创建隐藏窗口，用于接收剪贴板变化消息
	// 参数说明：扩展样式、类名、窗口名、样式、位置、大小、父窗口、菜单、实例、附加数据
	hwnd, _, err := winapi.ProcCreateWindowExW.Call(
		0,                                  // 扩展窗口样式，0表示默认
		uintptr(unsafe.Pointer(className)), // 窗口类名
		uintptr(unsafe.Pointer(className)), // 窗口名
		0,                                  // 窗口样式，0表示默认
		0, 0, 0, 0,                         // 窗口位置和大小，全0表示隐藏
		0,                          // 父窗口句柄，0表示桌面
		0,                          // 菜单句柄，0表示无菜单
		hInstance,                  // 应用程序实例句柄
		uintptr(unsafe.Pointer(a)), // 附加数据，传入应用实例指针
	)
	if hwnd == 0 {
		return fmt.Errorf("创建隐藏窗口失败: %v", err)
	}
	// 确保退出时销毁窗口，释放系统资源
	defer winapi.ProcDestroyWindow.Call(hwnd)

	// 注册剪贴板格式监听器，这样当剪贴板内容变化时会收到WM_CLIPBOARDUPDATE消息
	if ret, _, err := winapi.ProcAddClipboardFormatListener.Call(hwnd); ret == 0 {
		return fmt.Errorf("无法注册剪贴板监听: %v", err)
	}
	// 确保退出时取消注册剪贴板监听
	defer winapi.ProcRemoveClipboardFormatListener.Call(hwnd)

	// 启动一个goroutine监听退出信号，以便优雅地退出消息循环
	// 当收到信号或上下文被取消时，向消息循环发送退出消息
	go func(tid uint32) {
		select {
		case <-a.sigCh: // 收到操作系统退出信号（如Ctrl+C）
			postQuitToThread(tid)
		case <-a.ctx.Done(): // 上下文被取消（应用程序主动退出）
			postQuitToThread(tid)
		}
	}(tid)

	// 进入Windows消息循环，等待并处理各种系统消息
	var m Msg // 消息结构体，用于接收消息
	for {
		select {
		case <-a.ctx.Done():
			// 如果上下文被取消，退出消息循环
			return nil
		default:
			// 非阻塞式检查，继续处理消息
		}

		// 从消息队列中获取消息
		// 参数：消息结构体指针、窗口句柄过滤（0表示所有窗口）、消息范围过滤（0,0表示所有消息）
		ret, _, err := winapi.ProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			// 返回-1表示发生错误
			return fmt.Errorf("消息循环错误: %v", err)
		}
		if ret == 0 {
			// 返回0表示收到WM_QUIT消息，应该退出消息循环
			break
		}

		// 检查是否是剪贴板更新消息
		if m.Message == WMClipboardUpdate {
			// 调用剪贴板变化处理函数
			a.processClipboardChange()
		}

		// 将虚拟键消息转换为字符消息（如键盘输入）
		winapi.ProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		// 将消息分发给窗口过程函数进行处理
		winapi.ProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	return nil
}

// windowProc 处理窗口消息
// 这是窗口过程函数，负责处理系统发送给窗口的各种消息
// 参数:
//   - hwnd: 窗口句柄
//   - message: 消息类型
//   - wparam: 消息附加参数
//   - lparam: 消息附加参数
//
// 返回值:
//   - uintptr: 消息处理结果
func (a *PathConvertApp) windowProc(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	switch message {
	case WMDestroy:
		// 收到窗口销毁消息，向消息循环发送退出消息
		winapi.ProcPostQuitMessage.Call(0)
		return 0
	}
	// 对于未处理的消息，调用默认窗口过程函数
	ret, _, _ := winapi.ProcDefWindowProcW.Call(hwnd, uintptr(message), wparam, lparam)
	return ret
}

// runWithPolling 轮询模式
// 这是剪贴板监听API不可用时的备用实现，通过定期轮询检查剪贴板内容变化
// 工作原理:
//  1. 创建定时器，按照配置的间隔定期检查剪贴板
//  2. 比较当前剪贴板内容与上次记录的内容
//  3. 如果内容有变化，触发转换处理
//
// 返回值:
//   - error: 运行过程中可能发生的错误
func (a *PathConvertApp) runWithPolling() error {
	a.log.Info("使用轮询模式，间隔: %v", a.cfg.PollInterval)
	// 创建定时器，按照配置的时间间隔触发
	ticker := time.NewTicker(a.cfg.PollInterval)
	// 确保退出时停止定时器，防止资源泄漏
	defer ticker.Stop()

	// 进入轮询循环
	for {
		select {
		case <-ticker.C:
			// 定时器触发，检查剪贴板是否变化
			changed, err := a.cb.HasChanged()
			if err != nil {
				a.log.Debug("检查剪贴板时出错: %v", err)
				continue // 出错时继续下一次检查
			}
			if changed {
				// 剪贴板内容有变化，处理变化
				a.processClipboardChange()
			}
		case <-a.sigCh:
			// 收到退出信号
			a.log.Info("收到停止信号")
			if a.cancel != nil {
				// 取消上下文，通知所有协程退出
				a.cancel()
			}
			return nil
		case <-a.ctx.Done():
			// 上下文被取消，退出循环
			return nil
		}
	}
}

// processClipboardChange 处理剪贴板变化
// 这是剪贴板处理的核心函数，负责检查、转换并更新剪贴板内容
// 执行流程:
//  1. 检查自动转换是否启用
//  2. 获取当前剪贴板内容
//  3. 检查是否需要转换
//  4. 执行转换并更新剪贴板
func (a *PathConvertApp) processClipboardChange() {
	a.log.Debug("检测到剪贴板变化")
	// 检查用户是否禁用了自动转换功能
	if !a.cfg.AutoConvert {
		a.log.Debug("自动转换已禁用，忽略变化")
		return
	}

	// 获取剪贴板中的文本内容
	rawText, err := a.cb.GetText()
	if err != nil {
		a.log.Debug("无法获取剪贴板内容: %v", err)
		return
	}
	// 标准化文本内容，去除首尾空白字符
	normalized := strings.TrimSpace(rawText)

	// 计算当前内容的哈希值，用于快速比较内容是否变化
	currentHash := clipboard.QuickHash(normalized)
	// 检查内容是否真的发生了变化（避免重复处理）
	if currentHash == a.cb.LastContentHash() {
		a.log.Debug("剪贴板内容未变化，跳过处理")
		return
	}

	// 检查内容是否需要转换（路径转换器会判断内容是否包含Windows路径）
	if !a.pc.ShouldConvert(rawText) {
		a.log.Debug("不需要转换的内容: %s", a.log.ShortenText(rawText))
		// 更新最后处理的哈希值，避免下次重复检查
		a.cb.SetLastContentHash(currentHash)
		return
	}

	// 执行路径转换
	converted := a.pc.Convert(rawText)
	// 检查转换是否改变了内容（防止设置相同内容导致循环触发）
	if converted != rawText {
		// 将转换后的内容设置回剪贴板
		if err := a.cb.SetText(converted); err != nil {
			a.log.Error("无法设置剪贴板内容: %v", err)
			return
		}

		// 根据用户配置决定是否显示转换通知
		if a.cfg.ShowNotifications {
			a.log.Info("已转换路径:")
			a.log.Info("  原路径: %s", rawText)
			a.log.Info("  转换后: %s", converted)
		} else {
			a.log.Debug("已转换路径，但不显示通知")
		}

		// 更新最后处理的哈希值（使用转换后的内容的哈希）
		a.cb.SetLastContentHash(clipboard.QuickHash(converted))
		return
	}

	// 内容不需要转换，但更新哈希值以避免下次重复检查
	a.cb.SetLastContentHash(currentHash)
}

// getCurrentThreadID 获取当前线程ID
// 这是一个辅助函数，用于获取当前线程的ID，用于向特定线程发送消息
// 返回值:
//   - uint32: 当前线程的ID
func getCurrentThreadID() uint32 {
	// 调用Windows API获取当前线程ID
	ret, _, _ := winapi.ProcGetCurrentThreadId.Call()
	return uint32(ret)
}

// postQuitToThread 向指定线程发送退出消息
// 这是一个辅助函数，用于向指定线程发送WM_QUIT消息，结束其消息循环
// 参数:
//   - tid: 目标线程ID
func postQuitToThread(tid uint32) {
	// 向指定线程的消息队列发送退出消息
	// 参数：线程ID、消息类型（WM_QUIT）、wParam、lParam
	winapi.ProcPostThreadMessage.Call(uintptr(tid), uintptr(WMQuit), 0, 0)
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
