package clipboard

import (
	"testing"
)

func BenchmarkQuickHash_ShortText(b *testing.B) {
	text := "short text"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuickHash(text)
	}
}

func BenchmarkQuickHash_MediumText(b *testing.B) {
	text := "C:\\Users\\test\\Documents\\file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuickHash(text)
	}
}

func BenchmarkQuickHash_LongText(b *testing.B) {
	text := "C:\\Very\\Long\\Path\\With\\Many\\Directories\\And\\Files\\document.txt\\another\\level\\deep"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuickHash(text)
	}
}

func BenchmarkQuickHash_VeryLongText(b *testing.B) {
	text := "C:\\Program Files\\Application\\Very\\Long\\Directory\\Structure\\With\\Many\\Subdirectories\\And\\Files\\that\\go\\on\\and\\on\\config.ini"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuickHash(text)
	}
}
