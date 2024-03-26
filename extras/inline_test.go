package extras

import (
	"bytes"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func TestSuperscript(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(Superscript),
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

func TestSubscript(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Subscript, extension.Strikethrough,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/subscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSubscriptDump(t *testing.T) {
	input := "The H~2~O molecule"
	markdown := goldmark.New(
		goldmark.WithExtensions(Subscript),
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
		markdown := goldmark.New(goldmark.WithExtensions(Subscript))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestInsert(t *testing.T) {
	markdown := goldmark.New(goldmark.WithExtensions(Insert))
	testutil.DoTestCaseFile(markdown, "_test/insert.txt", t, testutil.ParseCliCaseArg()...)
}

func TestInsertDump(t *testing.T) {
	input := "Add some text: ++insertion++. Amazing."
	markdown := goldmark.New(goldmark.WithExtensions(Insert))
	root := markdown.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutInsert(b *testing.B) {
	const input = `
## Insert text explicitly

Add some text: ++insertion++. Amazing.`

	b.Run("without insert", func(b *testing.B) {
		markdown := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with insert", func(b *testing.B) {
		markdown := goldmark.New(goldmark.WithExtensions(Insert))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestMark(t *testing.T) {
	markdown := goldmark.New(goldmark.WithExtensions(Mark))
	testutil.DoTestCaseFile(markdown, "_test/mark.txt", t, testutil.ParseCliCaseArg()...)
}

func TestMarkDump(t *testing.T) {
	input := "Add some marked text: ==marked==. Amazing."
	markdown := goldmark.New(goldmark.WithExtensions(Mark))
	root := markdown.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutMark(b *testing.B) {
	const input = `
## Mark text

Add some marked text: ==marked==. Amazing.`

	b.Run("without mark extension", func(b *testing.B) {
		markdown := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with mark extension", func(b *testing.B) {
		markdown := goldmark.New(goldmark.WithExtensions(Mark))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
