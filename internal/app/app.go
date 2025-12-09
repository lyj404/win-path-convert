package app

import "win-path-convert/internal/winapi"

// Message and window-related constants reused by the application.
const (
	WMClipboardUpdate = winapi.WMClipboardUpdate
	WMDestroy         = winapi.WMDestroy
	WMQuit            = winapi.WMQuit
)

// WndClassEx mirrors the Windows WNDCLASSEX structure used to register a window class.
type WndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// Msg mirrors the Windows MSG structure for message loops.
type Msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}
