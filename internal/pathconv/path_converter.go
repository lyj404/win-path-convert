package pathconv

import (
	"regexp"
	"strings"
	"win-path-convert/internal/logger"
)

// envVarPattern 预编译的环境变量格式检测
var envVarPattern = regexp.MustCompile(`%[^%]+%`)

// PathConverter 处理路径检测和转换
type PathConverter struct {
	excludePatterns []string
	excludeRegexps  []*regexp.Regexp
	logger          *logger.Logger
}

// NewPathConverter 创建新的路径转换器
func NewPathConverter(excludePatterns []string, l *logger.Logger) *PathConverter {
	pc := &PathConverter{
		excludePatterns: excludePatterns,
		logger:          l,
	}
	pc.compileExcludePatterns()
	return pc
}

// compileExcludePatterns 预编译排除模式的正则表达式
func (pc *PathConverter) compileExcludePatterns() {
	for _, pattern := range pc.excludePatterns {
		regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
		regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
		regexPattern = "^" + regexPattern + "$"

		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			pc.logger.Warn("无法编译排除模式 '%s': %v", pattern, err)
			continue
		}
		pc.excludeRegexps = append(pc.excludeRegexps, regex)
	}
}

// ShouldConvert 判断是否应该转换给定的文本
func (pc *PathConverter) ShouldConvert(text string) bool {
	if text == "" {
		return false
	}

	trimmed := strings.Trim(text, "\"")

	if !strings.Contains(trimmed, "\\") {
		return false
	}

	if pc.isExcluded(trimmed) {
		return false
	}

	if len(trimmed) >= 3 {
		if trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/') {
			return true
		}
	}

	if strings.HasPrefix(trimmed, "\\\\") {
		return true
	}

	if strings.Contains(trimmed, "\\") {
		return true
	}

	return false
}

// isExcluded 检查文本是否匹配任何排除模式
func (pc *PathConverter) isExcluded(text string) bool {
	lowerText := strings.ToLower(text)

	if strings.HasPrefix(lowerText, "http://") || strings.HasPrefix(lowerText, "https://") {
		pc.logger.Debug("排除URL: %s", text)
		return true
	}

	if strings.HasPrefix(lowerText, "mailto:") ||
		strings.HasPrefix(lowerText, "ftp://") ||
		strings.HasPrefix(lowerText, "file://") {
		pc.logger.Debug("排除特殊协议: %s", text)
		return true
	}

	for _, regex := range pc.excludeRegexps {
		if regex.MatchString(text) {
			pc.logger.Debug("排除匹配模式的文本: %s", text)
			return true
		}
	}

	// 环境变量格式，仍允许转换其中的反斜杠
	if strings.Count(text, "%") >= 2 && envVarPattern.MatchString(text) {
		parts := strings.Split(text, "%")
		for i := 1; i < len(parts)-1; i += 2 {
			if strings.Contains(parts[i], "\\") {
				return true
			}
		}
	}

	return false
}

// Convert 将Windows路径转换为Unix风格路径
func (pc *PathConverter) Convert(text string) string {
	hasQuotes := strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`)
	content := strings.Trim(text, `"`)

	originalContent := content
	converted := strings.ReplaceAll(content, "\\", "/")

	if converted == originalContent {
		return text
	}

	if hasQuotes {
		converted = `"` + converted + `"`
	}

	pc.logger.Debug("路径转换: %s -> %s", originalContent, converted)
	return converted
}

// UpdateExcludePatterns 更新排除模式
func (pc *PathConverter) UpdateExcludePatterns(patterns []string) {
	pc.excludePatterns = patterns
	pc.excludeRegexps = nil
	pc.compileExcludePatterns()
}
