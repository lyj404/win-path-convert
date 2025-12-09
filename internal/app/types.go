package app

import "github.com/lyj404/win-path-convert/internal/winapi"

// Windows消息系统相关常量
// 这些常量定义了应用程序需要处理的Windows消息类型

// WMClipboardUpdate 表示剪贴板内容已更新的消息
// 当用户复制内容到剪贴板时，系统会向注册的窗口发送此消息
const (
	WMClipboardUpdate = winapi.WMClipboardUpdate // 剪贴板更新消息 (0x031D)
	WMDestroy         = winapi.WMDestroy         // 窗口销毁消息 (0x0002)
	WMQuit            = winapi.WMQuit            // 退出消息，用于结束消息循环 (0x0012)
)

// WndClassEx 窗口类结构体
// 这是Windows WNDCLASSEX结构的镜像，用于注册窗口类
// 窗口类定义了窗口的通用属性和行为，所有基于该类创建的窗口都会共享这些属性
type WndClassEx struct {
	CbSize        uint32  // 结构体大小，必须设置为sizeof(WNDCLASSEX)
	Style         uint32  // 窗口类样式，定义窗口类的行为特性
	LpfnWndProc   uintptr // 窗口过程函数指针，指向处理窗口消息的函数
	CbClsExtra    int32   // 额外的类内存，通常为0
	CbWndExtra    int32   // 额外的窗口内存，通常为0或用于存储窗口实例数据
	HInstance     uintptr // 应用程序实例句柄，标识加载窗口类的模块
	HIcon         uintptr // 类图标句柄，通常为0表示使用默认图标
	HCursor       uintptr // 类光标句柄，通常为0表示使用默认光标
	HbrBackground uintptr // 类背景画刷句柄，0表示使用默认背景
	LpszMenuName  *uint16 // 指向菜单资源名称字符串的指针，0表示无菜单
	LpszClassName *uint16 // 指向窗口类名称字符串的指针，必须唯一
	HIconSm       uintptr // 小图标句柄，用于任务栏和标题栏
}

// Msg 消息结构体
// 这是Windows MSG结构的镜像，用于存储从消息队列中检索的消息
// 消息循环会不断检索这些消息并分发给相应的窗口过程函数处理
type Msg struct {
	Hwnd    uintptr  // 消息目标窗口的句柄
	Message uint32   // 消息标识符，如WM_PAINT, WM_MOUSEMOVE等
	WParam  uintptr  // 消息的附加信息，具体含义取决于消息类型
	LParam  uintptr  // 消息的附加信息，具体含义取决于消息类型
	Time    uint32   // 消息投递到队列的时间
	Pt      struct { // 消息投递时鼠标光标的屏幕坐标
		X int32 // 屏幕X坐标
		Y int32 // 屏幕Y坐标
	}
}
