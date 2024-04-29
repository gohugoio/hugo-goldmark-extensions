package extras

import (
	"bytes"
	xast "github.com/gohugoio/hugo-goldmark-extensions/extras/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
)

func buildGoldmarkWithInlineTag(tag xast.InlineTagType) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(New(ExtraInlineTagConfig{
			InlineTagType: tag,
			Enable:        true,
		})))
}

var markdown = goldmark.New()
var markdownWithSuperscript = buildGoldmarkWithInlineTag(xast.Superscript)
var markdownWithSubscript = buildGoldmarkWithInlineTag(xast.Subscript)
var markdownWithInsert = buildGoldmarkWithInlineTag(xast.Insert)
var markdownWithMark = buildGoldmarkWithInlineTag(xast.Mark)

func TestSuperscript(t *testing.T) {
	testutil.DoTestCaseFile(markdownWithSuperscript, "_test/superscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSuperscriptDump(t *testing.T) {
	input := "Parabola: f(x) = x^2^. Amazing"
	root := markdownWithSuperscript.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
}

func BenchmarkWithAndWithoutOneSuperscript(b *testing.B) {
	const input = `
## Parabola

This formula contains one superscript: f(x) = x^2^ .`

	b.Run("without superscript", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with superscript", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdownWithSuperscript.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestSubscript(t *testing.T) {
	var markdown = goldmark.New(
		goldmark.WithExtensions(New(ExtraInlineTagConfig{
			InlineTagType: xast.Subscript,
			Enable:        true,
		}), extension.Strikethrough))
	testutil.DoTestCaseFile(markdown, "_test/subscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSubscriptDump(t *testing.T) {
	input := "The H~2~O molecule"
	root := markdownWithSubscript.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
}

func BenchmarkWithAndWithoutOneSubscript(b *testing.B) {
	const input = `
## Water formula
 
The chemical formula for water H~2~O contains one subscript.`

	b.Run("without subscript", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with subscript", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdownWithSubscript.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestInsert(t *testing.T) {
	testutil.DoTestCaseFile(markdownWithInsert, "_test/insert.txt", t, testutil.ParseCliCaseArg()...)
}

func TestInsertDump(t *testing.T) {
	input := "Add some text: ++insertion++. Amazing."
	root := markdownWithInsert.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutInsert(b *testing.B) {
	const input = `
## Insert text explicitly

Add some text: ++insertion++. Amazing.`

	b.Run("without insert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with insert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdownWithInsert.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestMark(t *testing.T) {
	testutil.DoTestCaseFile(markdownWithMark, "_test/mark.txt", t, testutil.ParseCliCaseArg()...)
}

func TestMarkDump(t *testing.T) {
	input := "Add some marked text: ==marked==. Amazing."
	root := markdownWithMark.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
}

func BenchmarkWithAndWithoutMark(b *testing.B) {
	const input = `
## Mark text

Add some marked text: ==marked==. Amazing.`

	b.Run("without mark extension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with mark extension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdownWithMark.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
