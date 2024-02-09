package superscript

import (
	"bytes"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func buildTestParser() goldmark.Markdown {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Superscript,
		),
	)
	return markdown
}

func TestSuperscript(t *testing.T) {
	markdown := buildTestParser()
	testutil.DoTestCaseFile(markdown, "_test/superscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestDump(t *testing.T) {
	input := "Parabola: f(x) = x^2^ . Amazing"
	md := buildTestParser()
	root := md.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutOneSuperscript(b *testing.B) {
	const input = `
## Parabola

This formula contains one superscript: f(x) = x^2^ .`

	b.Run("without superscript", func(b *testing.B) {
		md := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with superscript", func(b *testing.B) {
		md := goldmark.New(
			goldmark.WithExtensions(
				Superscript,
			),
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkWithAndWithoutThreeSuperscript(b *testing.B) {
	const input = `
## Pythagoras formula 

This formula contains three superscripts: a^2^ + b^2^=c^2^ .`

	b.Run("without superscript", func(b *testing.B) {
		md := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with superscript", func(b *testing.B) {
		md := goldmark.New(
			goldmark.WithExtensions(
				Superscript,
			),
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
