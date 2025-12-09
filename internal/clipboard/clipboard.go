package clipboard

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"win-path-convert/internal/winapi"
)

// ClipboardManager 封装剪贴板操作
type ClipboardManager struct {
	lastContentHash string
}

// NewClipboardManager 创建新的剪贴板管理器
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{}
}

// GetText 获取剪贴板文本内容，包含简单退避重试
func (cm *ClipboardManager) GetText() (string, error) {
	var lastErr error
	for _, delay := range []time.Duration{0, 15 * time.Millisecond, 30 * time.Millisecond} {
		if delay > 0 {
			time.Sleep(delay)
		}
		if ret, _, err := winapi.ProcOpenClipboard.Call(0); ret == 0 {
			lastErr = err
			continue
		}
		defer winapi.ProcCloseClipboard.Call()

		hData, _, _ := winapi.ProcGetClipboardData.Call(winapi.CFUnicodeText)
		if hData == 0 {
			return "", fmt.Errorf("剪贴板无文本内容")
		}

		size, _, _ := winapi.ProcGlobalSize.Call(hData)
		if size == 0 {
			return "", fmt.Errorf("无法获取剪贴板数据大小")
		}

		ptr, _, _ := winapi.ProcGlobalLock.Call(hData)
		if ptr == 0 {
			return "", fmt.Errorf("无法锁定剪贴板内存")
		}
		defer winapi.ProcGlobalUnlock.Call(hData)

		units := int(size / unsafe.Sizeof(uint16(0)))
		if units == 0 {
			return "", fmt.Errorf("剪贴板数据大小为零")
		}

		buffer := make([]uint16, units)
		winapi.ProcRtlMoveMemory.Call(
			uintptr(unsafe.Pointer(&buffer[0])),
			ptr,
			size,
		)

		text := syscall.UTF16ToString(buffer)
		return text, nil
	}

	return "", fmt.Errorf("无法打开剪贴板: %v", lastErr)
}

// SetText 设置剪贴板文本内容，包含简单退避重试
func (cm *ClipboardManager) SetText(text string) error {
	var lastErr error
	for _, delay := range []time.Duration{0, 15 * time.Millisecond, 30 * time.Millisecond} {
		if delay > 0 {
			time.Sleep(delay)
		}
		if ret, _, err := winapi.ProcOpenClipboard.Call(0); ret == 0 {
			lastErr = err
			continue
		}
		defer winapi.ProcCloseClipboard.Call()

		winapi.ProcEmptyClipboard.Call()

		utf16Text, err := windows.UTF16FromString(text)
		if err != nil {
			return fmt.Errorf("无法转换文本为UTF16: %v", err)
		}

		dataLen := uintptr(len(utf16Text) * int(unsafe.Sizeof(utf16Text[0])))
		hMem, _, _ := winapi.ProcGlobalAlloc.Call(winapi.GMEMMoveable, dataLen)
		if hMem == 0 {
			return fmt.Errorf("无法分配剪贴板内存")
		}

		ptr, _, _ := winapi.ProcGlobalLock.Call(hMem)
		if ptr == 0 {
			winapi.ProcGlobalFree.Call(hMem)
			return fmt.Errorf("无法锁定剪贴板内存")
		}

		winapi.ProcRtlMoveMemory.Call(
			ptr,
			uintptr(unsafe.Pointer(&utf16Text[0])),
			dataLen,
		)
		winapi.ProcGlobalUnlock.Call(hMem)

		ret, _, _ := winapi.ProcSetClipboardData.Call(winapi.CFUnicodeText, hMem)
		if ret == 0 {
			winapi.ProcGlobalFree.Call(hMem)
			return fmt.Errorf("无法设置剪贴板数据")
		}

		return nil
	}

	return fmt.Errorf("无法打开剪贴板: %v", lastErr)
}

// HasChanged 检查剪贴板内容是否已变化
func (cm *ClipboardManager) HasChanged() (bool, error) {
	text, err := cm.GetText()
	if err != nil {
		return false, err
	}

	currentHash := QuickHash(text)

	changed := currentHash != cm.lastContentHash

	if changed {

		cm.lastContentHash = currentHash

	}

	return changed, nil

}

// LastContentHash 返回最近一次内容的哈希
func (cm *ClipboardManager) LastContentHash() string {
	return cm.lastContentHash
}

// SetLastContentHash 设置最近一次内容的哈希
func (cm *ClipboardManager) SetLastContentHash(hash string) {
	cm.lastContentHash = hash
}

// quickHash 使用 FNV-1a 64bit，长度固定且碰撞概率低
func QuickHash(text string) string {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(text))
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}

// AddClipboardListener 添加剪贴板监听器
func (cm *ClipboardManager) AddClipboardListener(hwnd uintptr) error {
	ret, _, err := winapi.ProcAddClipboardFormatListener.Call(hwnd)
	if ret == 0 {
		return fmt.Errorf("无法添加剪贴板监听器: %v", err)
	}
	return nil
}

// RemoveClipboardListener 移除剪贴板监听器
func (cm *ClipboardManager) RemoveClipboardListener(hwnd uintptr) {
	winapi.ProcRemoveClipboardFormatListener.Call(hwnd)
}
