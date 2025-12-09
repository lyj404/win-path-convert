package pathconv

import (
	"regexp"
	"strings"

	"github.com/lyj404/win-path-convert/internal/logger"
)

// envVarPattern 预编译的环境变量格式检测正则表达式
// 用于识别Windows环境变量格式，如 %PATH%、%USERPROFILE% 等
// 环境变量需要特殊处理，因为它们可能包含需要转换的路径部分
var envVarPattern = regexp.MustCompile(`%[^%]+%`)

// PathConverter 处理路径检测和转换的核心结构体
// 该结构体封装了路径转换的逻辑，包括路径检测规则和排除模式
type PathConverter struct {
	excludePatterns []string         // 用户配置的排除模式列表，支持通配符
	excludeRegexps  []*regexp.Regexp // 编译后的排除模式正则表达式，用于高效匹配
	logger          *logger.Logger   // 日志记录器，用于输出转换过程中的信息
}

// NewPathConverter 创建新的路径转换器实例
// 该函数初始化一个PathConverter实例，并预编译用户配置的排除模式
// 参数:
//   - excludePatterns: 排除模式列表，用于排除不需要转换的内容
//   - l: 日志记录器，用于记录转换过程和错误信息
//
// 返回值:
//   - *PathConverter: 初始化完成的路径转换器实例
func NewPathConverter(excludePatterns []string, l *logger.Logger) *PathConverter {
	// 创建PathConverter实例
	pc := &PathConverter{
		excludePatterns: excludePatterns, // 存储用户配置的排除模式
		logger:          l,               // 存储日志记录器
	}
	// 预编译排除模式，提高后续匹配效率
	pc.compileExcludePatterns()
	return pc
}

// compileExcludePatterns 预编译排除模式的正则表达式
// 该函数将用户配置的通配符模式转换为正则表达式，并编译以提高匹配效率
// 通配符支持: * 匹配任意字符序列, . 匹配字面点字符
// 例如: "*.txt" 将转换为 "^.*\.txt$"
func (pc *PathConverter) compileExcludePatterns() {
	// 遍历所有用户配置的排除模式
	for _, pattern := range pc.excludePatterns {
		// 将通配符模式转换为正则表达式模式
		// . 转义为 \\. (匹配字面点字符)
		regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
		// * 转换为 .* (匹配任意字符序列)
		regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
		// 添加开始和结束锚点，确保完全匹配
		regexPattern = "^" + regexPattern + "$"

		// 编译正则表达式
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			// 编译失败，记录警告并跳过该模式
			pc.logger.Warn("无法编译排除模式 '%s': %v", pattern, err)
			continue
		}
		// 将编译后的正则表达式添加到列表中
		pc.excludeRegexps = append(pc.excludeRegexps, regex)
	}
}

// ShouldConvert 判断是否应该转换给定的文本
// 该函数通过一系列规则判断文本是否包含需要转换的Windows路径
// 包括检查反斜杠、驱动器字母格式、网络路径等，并考虑排除模式
// 参数:
//   - text: 要检查的文本
//
// 返回值:
//   - bool: 如果文本包含需要转换的Windows路径，返回true，否则返回false
func (pc *PathConverter) ShouldConvert(text string) bool {
	// 空文本不需要转换
	if text == "" {
		return false
	}

	// 去除文本两端的引号，Windows路径常被引号包围
	trimmed := strings.Trim(text, "\"")

	// 如果不包含反斜杠，则不可能是Windows路径，无需转换
	if !strings.Contains(trimmed, "\\") {
		return false
	}

	// 检查是否匹配任何排除模式
	if pc.isExcluded(trimmed) {
		return false
	}

	// 检查是否为绝对路径格式 (如 C:\ 或 C:/)
	// 前3个字符应为 驱动器字母 + 冒号 + 路径分隔符
	if len(trimmed) >= 3 {
		if trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/') {
			return true
		}
	}

	// 检查是否为UNC路径格式 (网络路径，以 \\ 开头)
	if strings.HasPrefix(trimmed, "\\\\") {
		return true
	}

	// 如果包含反斜杠，则视为需要转换的路径
	if strings.Contains(trimmed, "\\") {
		return true
	}

	// 不满足任何路径特征，不需要转换
	return false
}

// isExcluded 检查文本是否匹配任何排除模式
// 该函数根据一系列规则判断文本是否应该被排除，不进行路径转换
// 包括URL、特殊协议、环境变量和用户自定义模式
// 参数:
//   - text: 要检查的文本
//
// 返回值:
//   - bool: 如果文本匹配排除模式，返回true，否则返回false
func (pc *PathConverter) isExcluded(text string) bool {
	// 将文本转换为小写，用于不区分大小写的匹配
	lowerText := strings.ToLower(text)

	// 排除URL (http:// or https://)
	if strings.HasPrefix(lowerText, "http://") || strings.HasPrefix(lowerText, "https://") {
		pc.logger.Debug("排除URL: %s", text)
		return true
	}

	// 排除其他特殊协议 (mailto:, ftp:, file:)
	if strings.HasPrefix(lowerText, "mailto:") ||
		strings.HasPrefix(lowerText, "ftp://") ||
		strings.HasPrefix(lowerText, "file://") {
		pc.logger.Debug("排除特殊协议: %s", text)
		return true
	}

	// 检查是否匹配任何用户定义的排除模式
	for _, regex := range pc.excludeRegexps {
		if regex.MatchString(text) {
			pc.logger.Debug("排除匹配模式的文本: %s", text)
			return true
		}
	}

	// 特殊处理环境变量格式 (如 %USERPROFILE%\Documents)
	// 环境变量格式需要保留，但其中的路径部分可以转换
	if strings.Count(text, "%") >= 2 && envVarPattern.MatchString(text) {
		// 分割文本，获取环境变量部分
		parts := strings.Split(text, "%")
		// 检查每个环境变量部分是否包含反斜杠
		for i := 1; i < len(parts)-1; i += 2 {
			if strings.Contains(parts[i], "\\") {
				// 如果环境变量包含反斜杠，则允许转换
				return true
			}
		}
		// 环境变量不含路径部分，排除转换
		return false
	}

	// 不匹配任何排除模式，可以转换
	return false
}

// Convert 将Windows路径转换为Unix风格路径
// 该函数将文本中的反斜杠(\)替换为正斜杠(/)，保持原有的引号格式
// 注意: 该函数不会验证文本是否为有效路径，仅执行字符替换
// 参数:
//   - text: 要转换的文本
//
// 返回值:
//   - string: 转换后的文本，如果不需要转换则返回原文
func (pc *PathConverter) Convert(text string) string {
	// 检查并记录文本是否被引号包围
	hasQuotes := strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`)
	// 移除文本两端的引号，只处理内容部分
	content := strings.Trim(text, `"`)

	// 保存原始内容，用于比较是否发生了变化
	originalContent := content
	// 将所有反斜杠替换为正斜杠
	converted := strings.ReplaceAll(content, "\\", "/")

	// 如果没有变化，直接返回原文
	if converted == originalContent {
		return text
	}

	// 如果原文本有引号，为转换后的内容添加引号
	if hasQuotes {
		converted = `"` + converted + `"`
	}

	// 记录转换过程（调试级别）
	pc.logger.Debug("路径转换: %s -> %s", originalContent, converted)
	return converted
}

// UpdateExcludePatterns 更新排除模式
// 该函数允许运行时更新排除模式，常用于配置热更新
// 参数:
//   - patterns: 新的排除模式列表
func (pc *PathConverter) UpdateExcludePatterns(patterns []string) {
	// 更新排除模式列表
	pc.excludePatterns = patterns
	// 清除之前编译的正则表达式
	pc.excludeRegexps = nil
	// 重新编译新的排除模式
	pc.compileExcludePatterns()
}
