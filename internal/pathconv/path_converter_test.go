package pathconv

import (
	"testing"

	"win-path-convert/internal/config"
	"win-path-convert/internal/logger"
)

func newTestConverter() *PathConverter {

	cfg := config.DefaultConfig()

	l := logger.NewLogger(cfg.LogLevel)

	return NewPathConverter(cfg.ExcludePatterns, l)

}

func TestShouldConvert_DrivePath(t *testing.T) {
	pc := newTestConverter()
	if !pc.ShouldConvert(`C:\Users\test\file.txt`) {
		t.Fatalf("expected drive path to be convertible")
	}
}

func TestShouldConvert_UNCPath(t *testing.T) {
	pc := newTestConverter()
	if !pc.ShouldConvert(`\\server\share\file.txt`) {
		t.Fatalf("expected UNC path to be convertible")
	}
}

func TestShouldConvert_ExcludeUrl(t *testing.T) {
	pc := newTestConverter()
	if pc.ShouldConvert(`https://example.com/a\b`) {
		t.Fatalf("expected URL to be excluded")
	}
}

func TestShouldConvert_CustomExcludePattern(t *testing.T) {
	pc := NewPathConverter([]string{`*.tmp`}, logger.NewLogger("info"))
	if pc.ShouldConvert(`C:\a\b\c.tmp`) {
		t.Fatalf("expected custom exclude pattern to block conversion")
	}
}

func TestConvert_ReplacesBackslashes(t *testing.T) {
	pc := newTestConverter()
	out := pc.Convert(`C:\a\b\c`)
	if out != `C:/a/b/c` {
		t.Fatalf("expected converted path, got %q", out)
	}
}

func TestConvert_PreservesQuotes(t *testing.T) {
	pc := newTestConverter()
	out := pc.Convert(`"C:\a\b\c"`)
	if out != `"C:/a/b/c"` {
		t.Fatalf("expected quotes preserved, got %q", out)
	}
}
