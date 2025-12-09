package singleton

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

// 用于强制单实例运行的互斥量句柄和名称
var (
	mutexHandle windows.Handle
	mutexName   = "PathConvertToolMutex"
	isSingle    = false
)

// SetMutexName 允许在调用InitSingleton之前覆盖互斥量名称
func SetMutexName(name string) {
	if name != "" {
		mutexName = name
	}
}

// InitSingleton 尝试创建/打开命名互斥量
// 如果此实例拥有互斥量（即没有其他实例在运行），则返回true
func InitSingleton() bool {
	namePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return false
	}

	handle, err := windows.CreateMutex(nil, false, namePtr)
	if err != nil {
		return false
	}

	mutexHandle = handle
	lastErr := windows.GetLastError()
	if lastErr == syscall.Errno(syscall.ERROR_ALREADY_EXISTS) {
		return false
	}

	isSingle = true
	return true
}

// ReleaseSingleton 释放互斥量句柄
func ReleaseSingleton() bool {
	if !isSingle || mutexHandle == 0 {
		return true
	}

	err := windows.CloseHandle(mutexHandle)
	if err != nil && err != windows.ERROR_INVALID_HANDLE {
		return false
	}

	mutexHandle = 0
	isSingle = false
	return true
}

// IsSingleton 报告此进程是否获取了互斥量
func IsSingleton() bool {
	return isSingle
}

// CheckSingleton 尝试打开现有的互斥量以检测正在运行的实例
func CheckSingleton() (bool, error) {
	namePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return false, fmt.Errorf("cannot convert mutex name: %w", err)
	}

	handle, err := windows.OpenMutex(windows.SYNCHRONIZE, false, namePtr)
	if handle != 0 {
		defer windows.CloseHandle(handle)
	}

	if err != nil {
		// 打开失败，可能没有互斥量存在
		return true, nil
	}

	// 打开成功，另一个实例正在运行
	return false, nil
}
