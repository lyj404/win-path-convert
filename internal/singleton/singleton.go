package singleton

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

// Mutex handle and name used to enforce single instance.
var (
	mutexHandle windows.Handle
	mutexName   = "PathConvertToolMutex"
	isSingle    = false
)

// SetMutexName allows overriding the mutex name before InitSingleton is called.
func SetMutexName(name string) {
	if name != "" {
		mutexName = name
	}
}

// InitSingleton tries to create/open the named mutex.
// Returns true if this instance owns the mutex (i.e., no other instance running).
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

// ReleaseSingleton releases the mutex handle.
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

// IsSingleton reports whether this process acquired the mutex.
func IsSingleton() bool {
	return isSingle
}

// CheckSingleton attempts to open an existing mutex to detect running instances.
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
		// Open failed, likely no mutex exists.
		return true, nil
	}

	// Open succeeded, another instance is running.
	return false, nil
}
