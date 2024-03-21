package insert

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/text"
	"testing"
)

func buildTestParser() goldmark.Markdown {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			Extension,
		),
	)
	return markdown
}

func Test(t *testing.T) {
	markdown := buildTestParser()
	testutil.DoTestCaseFile(markdown, "testCases.txt", t, testutil.ParseCliCaseArg()...)
}

func TestDump(t *testing.T) {
	input := "Add some text: ++insertion++. Amazing."
	markdown := buildTestParser()
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
