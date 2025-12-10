package config

import "time"

// Config 包含应用程序的所有配置选项
// 这个结构体定义了应用程序运行所需的各种参数，通过修改这些参数
// 可以调整应用程序的行为，如转换策略、通知方式和日志详细程度等
type Config struct {
	PollInterval time.Duration // 轮询间隔（仅在无法使用剪贴板监听API时使用）
	// 当Windows剪贴板监听API不可用或权限不足时，应用程序会回退到轮询模式
	// 该值定义了两次检查剪贴板内容之间的时间间隔，值越小响应越快但CPU占用越高

	AutoConvert bool // 是否自动转换路径
	// 控制应用程序是否在检测到剪贴板内容变化时自动执行路径转换
	// 设为false时，应用程序会监听剪贴板变化但不会执行实际转换

	ShowNotifications bool // 是否显示转换通知
	// 控制当路径被转换时是否在日志和通知中显示详细信息
	// 设为false时仅记录调试信息，不在用户界面显示转换详情

	ExcludePatterns []string // 排除的模式列表
	// 定义不需要进行路径转换的内容模式，支持通配符匹配
	// 例如："*.exe", "http://*" 等，可以防止特定文件、URL等被错误转换

	LogLevel string // 日志级别: debug, info, warn, error
	// 控制日志输出的详细程度，不同级别输出不同数量的信息：
	// - debug: 最详细，包括所有内部操作和状态
	// - info: 适中，包括用户关心的操作和状态变化
	// - warn: 只包含警告和错误信息
	// - error: 只包含错误信息

	MutexName string // 互斥量名称，用于防止多个实例同时运行
	// Windows互斥锁名称，确保同一时间只有一个程序实例在运行
	// 不同程序应使用不同的互斥量名称，避免相互冲突
}

// DefaultConfig 返回应用程序的默认配置
// 该函数提供了应用程序的初始配置，这些值经过精心选择，
// 适合大多数用户的基本使用场景，同时保持了系统的高效运行
func DefaultConfig() *Config {
	return &Config{
		// 100毫秒的轮询间隔在响应速度和CPU占用之间取得了良好平衡
		// 既能快速响应剪贴板变化，又不会造成明显的系统负担
		PollInterval: 100 * time.Millisecond,

		// 默认启用自动转换，这是程序的主要功能
		// 用户复制路径时可以立即获得转换后的结果
		AutoConvert: true,

		// 默认显示转换通知，让用户知道转换操作已执行
		// 帮助用户理解程序的工作状态和转换结果
		ShowNotifications: true,

		// 默认排除所有URL和特殊协议，避免错误转换网络链接和协议内容
		// 这些模式不会被当作路径处理，防止破坏有用的URL和协议内容
		ExcludePatterns: []string{
			"http://*", "https://*", // 排除所有HTTP和HTTPS URL
			"mailto:*", "ftp://*", "file://*", // 排除其他特殊协议
		},

		// 默认使用info日志级别，提供适当的信息量
		// 既能跟踪程序运行状态，又不会产生过多日志噪音
		LogLevel: "info",

		// 默认互斥量名称，确保程序的单一实例运行
		// 如果需要同时运行多个版本或变体，应修改此名称
		MutexName: "PathConvertToolMutex",
	}
}
