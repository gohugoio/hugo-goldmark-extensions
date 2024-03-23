package extras

import (
	"bytes"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func TestSubscript(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Subscript, Superscript,
			extension.Strikethrough,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/subscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSubscriptDump(t *testing.T) {
	input := "The H~2~O molecule"
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Subscript, Superscript,
			extension.Strikethrough,
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
		markdown := goldmark.New(goldmark.WithExtensions(Subscript, extension.Strikethrough))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
