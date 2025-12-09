package winapi

import (
	"golang.org/x/sys/windows" // Windows平台特定的系统调用接口
)

// 系统动态链接库句柄
// 这些是Windows系统中常用的DLL，包含了各种系统API函数

var (
	User32   = windows.NewLazySystemDLL("user32.dll")   // 用户界面API，包括窗口、消息、剪贴板等
	Kernel32 = windows.NewLazySystemDLL("kernel32.dll") // 核心系统API，包括内存管理、进程线程等
)

// 剪贴板相关的Windows API函数
// 这些函数用于访问和操作Windows剪贴板，支持文本数据的读取和写入

var (
	// 剪贴板访问函数
	ProcOpenClipboard    = User32.NewProc("OpenClipboard")    // 打开剪贴板，以阻止其他应用程序修改剪贴板内容
	ProcCloseClipboard   = User32.NewProc("CloseClipboard")   // 关闭剪贴板，释放剪贴板访问权限
	ProcGetClipboardData = User32.NewProc("GetClipboardData") // 获取剪贴板数据句柄
	ProcEmptyClipboard   = User32.NewProc("EmptyClipboard")   // 清空剪贴板内容
	ProcSetClipboardData = User32.NewProc("SetClipboardData") // 设置剪贴板数据，将数据句柄传递给剪贴板

	// 内存操作函数
	ProcGlobalAlloc   = Kernel32.NewProc("GlobalAlloc")   // 从堆中分配内存，返回可移动的内存块句柄
	ProcGlobalLock    = Kernel32.NewProc("GlobalLock")    // 锁定内存块，返回指向内存数据的指针
	ProcGlobalUnlock  = Kernel32.NewProc("GlobalUnlock")  // 解锁内存块，使其他程序可以访问
	ProcGlobalFree    = Kernel32.NewProc("GlobalFree")    // 释放内存块，释放之前分配的内存
	ProcGlobalSize    = Kernel32.NewProc("GlobalSize")    // 获取内存块大小
	ProcRtlMoveMemory = Kernel32.NewProc("RtlMoveMemory") // 内存块复制，相当于C语言的memcpy函数

	// 剪贴板监听函数
	ProcAddClipboardFormatListener    = User32.NewProc("AddClipboardFormatListener")    // 注册剪贴板格式监听器
	ProcRemoveClipboardFormatListener = User32.NewProc("RemoveClipboardFormatListener") // 移除剪贴板格式监听器
)

// 窗口和消息处理相关的Windows API函数
// 这些函数用于创建窗口、处理消息以及与Windows消息系统交互

var (
	// 窗口类注册与管理
	ProcRegisterClassExW = User32.NewProc("RegisterClassExW") // 注册窗口类
	ProcUnregisterClassW = User32.NewProc("UnregisterClassW") // 注销窗口类

	// 窗口创建与销毁
	ProcCreateWindowExW = User32.NewProc("CreateWindowExW") // 创建扩展窗口
	ProcDestroyWindow   = User32.NewProc("DestroyWindow")   // 销毁窗口

	// 窗口过程和消息处理
	ProcDefWindowProcW   = User32.NewProc("DefWindowProcW")   // 默认窗口过程函数
	ProcGetMessageW      = User32.NewProc("GetMessageW")      // 从消息队列中获取消息
	ProcTranslateMessage = User32.NewProc("TranslateMessage") // 转换虚拟键消息为字符消息
	ProcDispatchMessageW = User32.NewProc("DispatchMessageW") // 分发消息给窗口过程函数
	ProcPostQuitMessage  = User32.NewProc("PostQuitMessage")  // 向消息队列发送退出消息

	// 线程消息处理
	ProcPostThreadMessage = User32.NewProc("PostThreadMessageW") // 向指定线程的消息队列发送消息

	// 系统模块与线程管理
	ProcGetModuleHandleW   = Kernel32.NewProc("GetModuleHandleW")   // 获取模块句柄
	ProcGetCurrentThreadId = Kernel32.NewProc("GetCurrentThreadId") // 获取当前线程ID
)

// Windows系统常量定义
// 这些常量是Windows API调用中常用的参数值

const (
	// 剪贴板格式常量
	CFUnicodeText = 13 // 剪贴板Unicode文本格式标识符

	// 内存分配标志常量
	GMEMMoveable = 0x0002 // 可移动内存标志，表示内存块可以在内存中移动

	// Windows消息常量
	WMClipboardUpdate = 0x031D // 剪贴板内容更新消息，当剪贴板内容变化时发送
	WMDestroy         = 0x0002 // 窗口销毁消息，当窗口即将被销毁时发送
	WMQuit            = 0x0012 // 退出消息，用于请求消息循环终止
)
