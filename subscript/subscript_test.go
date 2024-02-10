package subscript

import (
	"bytes"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func buildTestParser() goldmark.Markdown {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Extension, extension.Strikethrough,
		),
	)
	return markdown
}

func Test(t *testing.T) {
	markdown := buildTestParser()
	testutil.DoTestCaseFile(markdown, "testCases.txt", t, testutil.ParseCliCaseArg()...)
}

func TestDump(t *testing.T) {
	input := "**H~2~O** or simply water."
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Extension,
		),
	)
	root := markdown.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutOneSubscript(b *testing.B) {
	const input = `
## Water formula
 
The chemical formula for water H~2~O contains one subscript.`

	b.Run("without subscript", func(b *testing.B) {
		markdown := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with subscript", func(b *testing.B) {
		markdown := buildTestParser()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
