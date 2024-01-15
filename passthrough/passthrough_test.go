package passthrough

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"

	qt "github.com/frankban/quicktest"
)

func buildTestParser() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(NewPassthroughWithDelimiters(
			/*inlines*/ []delimiters{
				{
					Open:  "$",
					Close: "$",
				},
				{
					Open:  "\\(",
					Close: "\\)",
				},
			},
			/*blocks*/ []delimiters{
				{
					Open:  "$$",
					Close: "$$",
				},
				{
					Open:  "\\[",
					Close: "\\[",
				},
			},
		)),
	)
	return md
}

func Parse(t *testing.T, input string) string {
	md := buildTestParser()
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(buf.String())
}

func TestEmphasisOutsideOfMathmode(t *testing.T) {
	input := "Emph: _wow_"
	expected := "<p>Emph: <em>wow</em></p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestInlineEquationWithEmphasisDelimiters(t *testing.T) {
	input := "An equation: $a^*=x-b^*$. Amazing"
	expected := "<p>An equation: $a^*=x-b^*$. Amazing</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestInlineEquationWithMultiByteDelimiters(t *testing.T) {
	input := "An equation: \\(a^*=x-b^*\\). Amazing"
	expected := "<p>An equation: \\(a^*=x-b^*\\). Amazing</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestInlineEquationWithEmphasisDelimitersSplitAcrossLines(t *testing.T) {
	input := `An equation: $a^*=
x-b^*$. Amazing`
	expected := `<p>An equation: $a^*=
x-b^*$. Amazing</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestInlineEquationWithEmphasisSplitAcrossParagraphs(t *testing.T) {
	input := `An equation: $a^

*=x-b^*$. Amazing`
	expected := `<p>An equation: $a^</p>
<p><em>=x-b^</em>$. Amazing</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestBlockEquationWithEmphasisDelimiters(t *testing.T) {
	input := `An equation:

$$
a^*=x-b^*
$$

Amazing`
	expected := `<p>An equation:</p>

$$
a^*=x-b^*
$$

<p>Amazing</p>`

	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestBlockEquationWithOpenAndCloseOnSameLines(t *testing.T) {
	input := `An equation:

$$a^*=x-b^*
=c$$

Amazing`
	expected := `<p>An equation:</p>

$$a^*=x-b^*
=c$$

<p>Amazing</p>`

	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestBlockEquationBreakingParagraph(t *testing.T) {
  input := `An equation: \\[a^*=x-b^*\\] Amazing.`
	// This one is treated as inline because, for whatever reason, the block
	// parser is never triggered, even though we set CanInterruptParagraph to be
	// true. Hence it does not trigger and gets mangled as normal.
	expected := `<p>An equation: \[a^<em>=x-b^</em>\] Amazing.</p>`

	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}
