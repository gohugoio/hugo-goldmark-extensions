package extras_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/gohugoio/hugo-goldmark-extensions/extras/v2"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

// md bundles a parser and renderer — the v2 replacement for goldmark.Markdown,
// which no longer exists.
type md struct {
	p parser.Parser
	r renderer.Renderer[io.Writer]
}

func newMD(parserExts []parser.Extension, htmlExts []html.Extension) md {
	return md{
		p: parser.New(parser.WithExtensions(parserExts...)),
		r: html.New(html.WithExtensions(htmlExts...)),
	}
}

// inlineMD builds a converter with the extras inline-tag extension for the given config.
func inlineMD(conf extras.Config) md {
	return newMD(
		[]parser.Extension{extras.NewParser(conf)},
		[]html.Extension{extras.NewHTMLRenderer(conf)},
	)
}

func (m md) Convert(source []byte, w io.Writer) error {
	return m.r.Render(w, source, m.p.Parse(source))
}

func (m md) stringFunc() testutil.MarkdownToStringFunc {
	return testutil.NewMarkdownToStringFunc(m.p, m.r)
}

var (
	markdown                = newMD(nil, nil)
	markdownWithSuperscript = inlineMD(extras.Config{Superscript: extras.SuperscriptConfig{Enable: true}})
	markdownWithSubscript   = inlineMD(extras.Config{Subscript: extras.SubscriptConfig{Enable: true}})
	markdownWithInsert      = inlineMD(extras.Config{Insert: extras.InsertConfig{Enable: true}})
	markdownWithMark        = inlineMD(extras.Config{Mark: extras.MarkConfig{Enable: true}})
	markdownWithDelete      = inlineMD(extras.Config{Delete: extras.DeleteConfig{Enable: true}})
	// Two separate extension instances, to check they compose.
	markdownWithDeleteAndSubscript = newMD(
		[]parser.Extension{
			extras.NewParser(extras.Config{Subscript: extras.SubscriptConfig{Enable: true}}),
			extras.NewParser(extras.Config{Delete: extras.DeleteConfig{Enable: true}}),
		},
		[]html.Extension{
			extras.NewHTMLRenderer(extras.Config{Subscript: extras.SubscriptConfig{Enable: true}}),
			extras.NewHTMLRenderer(extras.Config{Delete: extras.DeleteConfig{Enable: true}}),
		},
	)
)

// dump exercises a node's Dump method (via PrettyPrint, which recurses); it just
// must not crash.
func dump(m md, input string) {
	root := m.p.Parse([]byte(input))
	_ = root.Dump([]byte(input)).PrettyPrint(io.Discard, []byte(input))
}

func TestSuperscript(t *testing.T) {
	testutil.DoTestCaseFile(markdownWithSuperscript.stringFunc(), "_test/superscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSuperscriptDump(t *testing.T) {
	dump(markdownWithSuperscript, "Parabola: f(x) = x^2^. Amazing")
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
	testutil.DoTestCaseFile(markdownWithDeleteAndSubscript.stringFunc(), "_test/subscript.txt", t, testutil.ParseCliCaseArg()...)
}

func TestSubscriptDump(t *testing.T) {
	dump(markdownWithSubscript, "The H~2~O molecule")
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
	testutil.DoTestCaseFile(markdownWithInsert.stringFunc(), "_test/insert.txt", t, testutil.ParseCliCaseArg()...)
}

func TestInsertDump(t *testing.T) {
	dump(markdownWithInsert, "Add some text: ++insertion++. Amazing.")
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
	testutil.DoTestCaseFile(markdownWithMark.stringFunc(), "_test/mark.txt", t, testutil.ParseCliCaseArg()...)
}

func TestMarkDump(t *testing.T) {
	dump(markdownWithMark, "Add some marked text: ==marked==. Amazing.")
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

func TestDelete(t *testing.T) {
	testutil.DoTestCaseFile(markdownWithDelete.stringFunc(), "_test/delete.txt", t, testutil.ParseCliCaseArg()...)
}

func TestDeleteDump(t *testing.T) {
	dump(markdownWithDelete, "Delete some text: ~~deleted~~. Amazing.")
}

func BenchmarkWithAndWithoutDelete(b *testing.B) {
	const input = `
## Delete text

Delete some text: ~~deleted~~. Amazing.`

	b.Run("without delete extension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with delete extension", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := markdownWithDelete.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
