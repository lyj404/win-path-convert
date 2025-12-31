package app

import (
	"github.com/lyj404/win-path-convert/internal/clipboard"
)

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

	// 计算当前内容的哈希值，用于快速比较内容是否变化
	// 提前计算一次哈希，避免后续重复计算
	currentHash := clipboard.QuickHash(rawText)
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

		// 更新最后处理的哈希值（复用已计算的转换后内容的哈希，避免重复计算）
		convertedHash := clipboard.QuickHash(converted)
		a.cb.SetLastContentHash(convertedHash)
		return
	}

	// 内容不需要转换，但更新哈希值以避免下次重复检查
	a.cb.SetLastContentHash(currentHash)
}
