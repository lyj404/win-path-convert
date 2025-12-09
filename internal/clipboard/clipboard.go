package clipboard

import (
	"encoding/hex" // 用于将哈希值转换为十六进制字符串
	"fmt"          // 格式化输出
	"hash/fnv"     // FNV哈希算法实现
	"syscall"      // 系统调用接口
	"time"         // 时间操作，用于退避重试间隔
	"unsafe"       // 不安全指针操作，用于Windows API调用

	"golang.org/x/sys/windows" // Windows平台特定的系统调用

	"github.com/lyj404/win-path-convert/internal/winapi" // 内部Windows API封装
)

// ClipboardManager 封装剪贴板操作
// 这个结构体管理剪贴板的读写操作，并跟踪最近一次处理的内容哈希
// 通过内容哈希比较，可以快速检测剪贴板内容是否发生变化
type ClipboardManager struct {
	lastContentHash string // 最近一次剪贴板内容的哈希值，用于内容变化检测
}

// NewClipboardManager 创建新的剪贴板管理器
// 返回一个初始化的ClipboardManager实例，可以立即使用
// 返回值:
//   - *ClipboardManager: 新创建的剪贴板管理器实例
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{}
}

// GetText 获取剪贴板文本内容，包含简单退避重试
// 该函数尝试从Windows剪贴板中获取文本内容，如果剪贴板被其他进程占用，
// 会使用退避策略重试几次，提高获取数据的成功率
// 返回值:
//   - string: 剪贴板中的文本内容
//   - error: 获取过程中可能发生的错误
func (cm *ClipboardManager) GetText() (string, error) {
	var lastErr error // 记录最后一次错误，用于返回

	// 使用退避策略重试，第一次立即尝试，然后等待15ms和30ms再尝试
	// 这种策略可以减少因剪贴板被其他进程临时占用而导致的失败
	for _, delay := range []time.Duration{0, 15 * time.Millisecond, 30 * time.Millisecond} {
		if delay > 0 {
			time.Sleep(delay)
		}

		// 尝试打开剪贴板，0表示当前进程
		// 如果返回值非0表示成功，0表示失败
		if ret, _, err := winapi.ProcOpenClipboard.Call(0); ret == 0 {
			lastErr = err
			continue // 打开失败，尝试下一次重试
		}
		// 确保函数退出时关闭剪贴板，避免资源锁定
		defer winapi.ProcCloseClipboard.Call()

		// 获取剪贴板数据句柄，CF_UNICODE_TEXT表示Unicode文本格式
		hData, _, _ := winapi.ProcGetClipboardData.Call(winapi.CFUnicodeText)
		if hData == 0 {
			return "", fmt.Errorf("剪贴板无文本内容")
		}

		// 获取数据块的大小（以字节为单位）
		size, _, _ := winapi.ProcGlobalSize.Call(hData)
		if size == 0 {
			return "", fmt.Errorf("无法获取剪贴板数据大小")
		}

		// 锁定内存块，获取指向数据的指针
		// GlobalLock返回一个指向内存块的指针，用于读取数据
		ptr, _, _ := winapi.ProcGlobalLock.Call(hData)
		if ptr == 0 {
			return "", fmt.Errorf("无法锁定剪贴板内存")
		}
		// 确保函数退出时解锁内存块
		defer winapi.ProcGlobalUnlock.Call(hData)

		// 计算Unicode字符的数量（每个字符占2字节）
		units := int(size / unsafe.Sizeof(uint16(0)))
		if units == 0 {
			return "", fmt.Errorf("剪贴板数据大小为零")
		}

		// 创建缓冲区，用于存储Unicode字符
		buffer := make([]uint16, units)
		// 将剪贴板数据复制到缓冲区
		// RtlMoveMemory相当于C语言的memcpy函数，用于内存块复制
		winapi.ProcRtlMoveMemory.Call(
			uintptr(unsafe.Pointer(&buffer[0])),
			ptr,
			size,
		)

		// 将UTF-16编码的字符串转换为Go字符串
		text := syscall.UTF16ToString(buffer)
		return text, nil
	}

	// 所有尝试均失败，返回最后一次错误
	return "", fmt.Errorf("无法打开剪贴板: %v", lastErr)
}

// SetText 设置剪贴板文本内容，包含简单退避重试
// 该函数尝试将文本内容设置到Windows剪贴板中，使用退避策略重试几次，
// 提高设置数据的成功率
// 参数:
//   - text: 要设置的文本内容
//
// 返回值:
//   - error: 设置过程中可能发生的错误
func (cm *ClipboardManager) SetText(text string) error {
	var lastErr error // 记录最后一次错误，用于返回

	// 使用退避策略重试，与GetText相同的策略
	for _, delay := range []time.Duration{0, 15 * time.Millisecond, 30 * time.Millisecond} {
		if delay > 0 {
			time.Sleep(delay)
		}

		// 尝试打开剪贴板
		if ret, _, err := winapi.ProcOpenClipboard.Call(0); ret == 0 {
			lastErr = err
			continue // 打开失败，尝试下一次重试
		}
		// 确保函数退出时关闭剪贴板
		defer winapi.ProcCloseClipboard.Call()

		// 清空剪贴板，准备设置新内容
		winapi.ProcEmptyClipboard.Call()

		// 将Go字符串转换为UTF-16编码的字节切片
		utf16Text, err := windows.UTF16FromString(text)
		if err != nil {
			return fmt.Errorf("无法转换文本为UTF16: %v", err)
		}

		// 计算数据长度（以字节为单位）
		dataLen := uintptr(len(utf16Text) * int(unsafe.Sizeof(utf16Text[0])))
		// 分配可移动的内存块，用于存储剪贴板数据
		// GMEM_MOVEABLE表示内存块可以被移动和重新分配
		hMem, _, _ := winapi.ProcGlobalAlloc.Call(winapi.GMEMMoveable, dataLen)
		if hMem == 0 {
			return fmt.Errorf("无法分配剪贴板内存")
		}

		// 锁定内存块，获取指向数据的指针
		ptr, _, _ := winapi.ProcGlobalLock.Call(hMem)
		if ptr == 0 {
			// 锁定失败，释放已分配的内存
			winapi.ProcGlobalFree.Call(hMem)
			return fmt.Errorf("无法锁定剪贴板内存")
		}

		// 将UTF-16数据复制到分配的内存块中
		winapi.ProcRtlMoveMemory.Call(
			ptr,
			uintptr(unsafe.Pointer(&utf16Text[0])),
			dataLen,
		)
		// 解锁内存块，使其可以被剪贴板访问
		winapi.ProcGlobalUnlock.Call(hMem)

		// 设置剪贴板数据，数据句柄的所有权转移给剪贴板系统
		ret, _, _ := winapi.ProcSetClipboardData.Call(winapi.CFUnicodeText, hMem)
		if ret == 0 {
			// 设置失败，释放内存块
			winapi.ProcGlobalFree.Call(hMem)
			return fmt.Errorf("无法设置剪贴板数据")
		}

		// 设置成功，返回nil
		return nil
	}

	// 所有尝试均失败，返回最后一次错误
	return fmt.Errorf("无法打开剪贴板: %v", lastErr)
}

// HasChanged 检查剪贴板内容是否已变化
// 该函数通过比较当前剪贴板内容的哈希值与上次记录的哈希值，
// 判断剪贴板内容是否发生变化，避免对相同内容的重复处理
// 返回值:
//   - bool: 内容是否发生变化
//   - error: 检查过程中可能发生的错误
func (cm *ClipboardManager) HasChanged() (bool, error) {
	// 获取当前剪贴板内容
	text, err := cm.GetText()
	if err != nil {
		return false, err
	}

	// 计算当前内容的哈希值
	currentHash := QuickHash(text)

	// 比较当前哈希值与上次记录的哈希值
	changed := currentHash != cm.lastContentHash

	// 如果内容发生变化，更新记录的哈希值
	if changed {
		cm.lastContentHash = currentHash
	}

	return changed, nil
}

// LastContentHash 返回最近一次内容的哈希
// 该函数返回上次记录的剪贴板内容哈希值，主要用于比较操作
// 返回值:
//   - string: 最近一次内容的哈希值
func (cm *ClipboardManager) LastContentHash() string {
	return cm.lastContentHash
}

// SetLastContentHash 设置最近一次内容的哈希
// 该函数用于手动设置上次记录的剪贴板内容哈希值，
// 主要用于初始化或特殊情况下的哈希值更新
// 参数:
//   - hash: 要设置的哈希值
func (cm *ClipboardManager) SetLastContentHash(hash string) {
	cm.lastContentHash = hash
}

// QuickHash 使用 FNV-1a 64bit，长度固定且碰撞概率低
// 该函数计算文本内容的哈希值，用于快速比较两段文本是否相同
// 使用FNV-1a哈希算法，该算法计算速度快，碰撞概率低
// 参数:
//   - text: 要计算哈希的文本
//
// 返回值:
//   - string: 文本的十六进制哈希值
func QuickHash(text string) string {
	// 创建FNV-1a 64位哈希器
	hasher := fnv.New64a()
	// 将文本写入哈希器
	_, _ = hasher.Write([]byte(text))
	// 计算哈希值
	sum := hasher.Sum(nil)
	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(sum)
}

// AddClipboardListener 添加剪贴板监听器
// 该函数将指定窗口注册为剪贴板格式监听器，当剪贴板内容发生变化时，
// 系统会向该窗口发送WM_CLIPBOARDUPDATE消息
// 参数:
//   - hwnd: 要注册的窗口句柄
//
// 返回值:
//   - error: 注册过程中可能发生的错误
func (cm *ClipboardManager) AddClipboardListener(hwnd uintptr) error {
	// 调用Windows API注册剪贴板格式监听器
	ret, _, err := winapi.ProcAddClipboardFormatListener.Call(hwnd)
	if ret == 0 {
		return fmt.Errorf("无法添加剪贴板监听器: %v", err)
	}
	return nil
}

// RemoveClipboardListener 移除剪贴板监听器
// 该函数取消指定窗口的剪贴板格式监听器注册，停止接收剪贴板变化通知
// 参数:
//   - hwnd: 要取消注册的窗口句柄
func (cm *ClipboardManager) RemoveClipboardListener(hwnd uintptr) {
	// 调用Windows API移除剪贴板格式监听器
	winapi.ProcRemoveClipboardFormatListener.Call(hwnd)
}
