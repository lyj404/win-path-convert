package pathconv

import (
	"github.com/lyj404/win-path-convert/internal/config"
	"github.com/lyj404/win-path-convert/internal/logger"
	"testing"
)

func BenchmarkShouldConvert_DrivePath(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testPath := `C:\Users\test\Documents\file.txt`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.ShouldConvert(testPath)
	}
}

func BenchmarkShouldConvert_UNCPath(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testPath := `\\server\share\folder\file.txt`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.ShouldConvert(testPath)
	}
}

func BenchmarkShouldConvert_URL(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testURL := `https://example.com/path/to/file`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.ShouldConvert(testURL)
	}
}

func BenchmarkShouldConvert_WithExclusions(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter([]string{"*.tmp", "*.log"}, l)

	testPath := `C:\test\file.txt`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.ShouldConvert(testPath)
	}
}

func BenchmarkConvert_SimplePath(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testPath := `C:\Users\test\Documents`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Convert(testPath)
	}
}

func BenchmarkConvert_LongPath(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testPath := `C:\Very\Long\Path\With\Many\Subdirectories\And\Files\document.txt`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Convert(testPath)
	}
}

func BenchmarkConvert_WithQuotes(b *testing.B) {
	cfg := config.DefaultConfig()
	l := logger.NewLogger(cfg.LogLevel)
	pc := NewPathConverter(cfg.ExcludePatterns, l)

	testPath := `"C:\Program Files\Application\config.ini"`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Convert(testPath)
	}
}
