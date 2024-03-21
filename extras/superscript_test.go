package extras

import (
	"bytes"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func TestSuperscript(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Superscript, Subscript,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/superscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSuperscriptDump(t *testing.T) {
	input := "Parabola: f(x) = x^2^. Amazing"
	markdown := goldmark.New(goldmark.WithExtensions(Superscript))
	root := markdown.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutOneSuperscript(b *testing.B) {
	const input = `
## Parabola

This formula contains one superscript: f(x) = x^2^ .`

	b.Run("without superscript", func(b *testing.B) {
		markdown := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with superscript", func(b *testing.B) {
		markdown := goldmark.New(goldmark.WithExtensions(Superscript))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
