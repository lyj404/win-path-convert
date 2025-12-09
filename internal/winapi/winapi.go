package winapi

import (
	"golang.org/x/sys/windows"
)

var (
	User32   = windows.NewLazySystemDLL("user32.dll")
	Kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// 常用剪贴板函数
	ProcOpenClipboard    = User32.NewProc("OpenClipboard")
	ProcCloseClipboard   = User32.NewProc("CloseClipboard")
	ProcGetClipboardData = User32.NewProc("GetClipboardData")
	ProcEmptyClipboard   = User32.NewProc("EmptyClipboard")
	ProcSetClipboardData = User32.NewProc("SetClipboardData")

	// 内存操作
	ProcGlobalAlloc   = Kernel32.NewProc("GlobalAlloc")
	ProcGlobalLock    = Kernel32.NewProc("GlobalLock")
	ProcGlobalUnlock  = Kernel32.NewProc("GlobalUnlock")
	ProcGlobalFree    = Kernel32.NewProc("GlobalFree")
	ProcGlobalSize    = Kernel32.NewProc("GlobalSize")
	ProcRtlMoveMemory = Kernel32.NewProc("RtlMoveMemory")

	// 剪贴板监听
	ProcAddClipboardFormatListener    = User32.NewProc("AddClipboardFormatListener")
	ProcRemoveClipboardFormatListener = User32.NewProc("RemoveClipboardFormatListener")

	// 窗口/消息
	ProcRegisterClassExW   = User32.NewProc("RegisterClassExW")
	ProcUnregisterClassW   = User32.NewProc("UnregisterClassW")
	ProcCreateWindowExW    = User32.NewProc("CreateWindowExW")
	ProcDestroyWindow      = User32.NewProc("DestroyWindow")
	ProcDefWindowProcW     = User32.NewProc("DefWindowProcW")
	ProcGetMessageW        = User32.NewProc("GetMessageW")
	ProcTranslateMessage   = User32.NewProc("TranslateMessage")
	ProcDispatchMessageW   = User32.NewProc("DispatchMessageW")
	ProcPostQuitMessage    = User32.NewProc("PostQuitMessage")
	ProcPostThreadMessage  = User32.NewProc("PostThreadMessageW")
	ProcGetModuleHandleW   = Kernel32.NewProc("GetModuleHandleW")
	ProcGetCurrentThreadId = Kernel32.NewProc("GetCurrentThreadId")
)

const (
	CFUnicodeText     = 13
	GMEMMoveable      = 0x0002
	WMClipboardUpdate = 0x031D
	WMDestroy         = 0x0002
	WMQuit            = 0x0012
)
