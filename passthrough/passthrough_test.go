package passthrough

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"

	qt "github.com/frankban/quicktest"
)

func buildTestParser() goldmark.Markdown {
	md := goldmark.New(
		goldmark.WithExtensions(New(
			Config{
				InlineDelimiters: []Delimiters{
					{
						Open:  "$",
						Close: "$",
					},
					{
						Open:  "\\(",
						Close: "\\)",
					},
				},
				BlockDelimiters: []Delimiters{
					{
						Open:  "$$",
						Close: "$$",
					},
					{
						Open:  "\\[",
						Close: "\\]",
					},
				},
			},
		)),
	)
	return md
}

func Parse(t *testing.T, input string) string {
	md := buildTestParser()
	var buf bytes.Buffer

	// root := md.Parser().Parse(text.NewReader([]byte(input)))
	// root.Dump([]byte(input), 0)

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

func TestDump(t *testing.T) {
	input := "An equation: \\(a^*=x-b^*\\). Amazing"
	md := buildTestParser()
	root := md.Parser().Parse(text.NewReader([]byte(input)))
	root.Dump([]byte(input), 0)
	// Prints to stdout, so just test that it doesn't crash
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

func TestInlineEquationWithEmphasisDelimitersSplitAcrossLines2(t *testing.T) {
	input := `Inline $
a^*=x-b^*
$ equation`
	expected := `<p>Inline $
a^*=x-b^*
$ equation</p>`
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

func TestUnterminatedDelimiters(t *testing.T) {
	input := `An equation: $a^*=x-b^* Amazing.`
	expected := `<p>An equation: $a^<em>=x-b^</em> Amazing.</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestEscapedSingleByteDelimiter(t *testing.T) {
	input := `I want \\$ *dollars*: $a^*=x-b^*$ Amazing.`
	expected := `<p>I want \$ <em>dollars</em>: $a^*=x-b^*$ Amazing.</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestFirstByteOfMultiByteDelimiterEndsText(t *testing.T) {
	input := `An equation: \`
	expected := `<p>An equation: \</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample1(t *testing.T) {
	input := `Inline $x = {-b \pm \sqrt{b^2-4ac} \over 2a}$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample2(t *testing.T) {
	input := `Inline $
x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample3(t *testing.T) {
	input := `Inline $x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample4(t *testing.T) {
	input := `Inline $
x = {-b \pm \sqrt{b^2-4ac} \over 2a}$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample5(t *testing.T) {
	input := `Block $$x = {-b \pm \sqrt{b^2-4ac} \over 2a}$$ equation`
	expected := `<p>Block </p>
$$x = {-b \pm \sqrt{b^2-4ac} \over 2a}$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample6(t *testing.T) {
	input := `Block $$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$ equation`
	expected := `<p>Block </p>
$$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample7(t *testing.T) {
	input := `Block $$x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$ equation`
	expected := `<p>Block </p>
$$x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$
<p> equation</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample8(t *testing.T) {
	input := `Block $$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}$$ equation`
	expected := `<p>Block </p>
$$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}$$
<p> equation</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample9(t *testing.T) {
	input := `Inline $a^*=x-b^*$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample10(t *testing.T) {
	input := `Inline $
a^*=x-b^*
$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample11(t *testing.T) {
	input := `Inline $a^*=x-b^*
$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample12(t *testing.T) {
	input := `Inline $
a^*=x-b^*$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample13(t *testing.T) {
	input := `Block $$a^*=x-b^*$$ equation`
	expected := `<p>Block </p>
$$a^*=x-b^*$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample14(t *testing.T) {
	input := `Block $$
a^*=x-b^*
$$ equation`
	expected := `<p>Block </p>
$$
a^*=x-b^*
$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample15(t *testing.T) {
	input := `Block $$a^*=x-b^*
$$ equation`
	expected := `<p>Block </p>
$$a^*=x-b^*
$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample16(t *testing.T) {
	input := `Block $$
a^*=x-b^*$$ equation`
	expected := `<p>Block </p>
$$
a^*=x-b^*$$
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample17(t *testing.T) {
	input := `Inline \(a^*=x-b^*\) equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample18(t *testing.T) {
	input := `Inline \(
a^*=x-b^*
\) equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample19(t *testing.T) {
	input := `Inline \(a^*=x-b^*
\) equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample20(t *testing.T) {
	input := `Inline \(
a^*=x-b^*\) equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample21(t *testing.T) {
	input := `Block \[a^*=x-b^*\] equation`
	expected := `<p>Block </p>
\[a^*=x-b^*\]
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample22(t *testing.T) {
	input := `Block \[
a^*=x-b^*
\] equation`
	expected := `<p>Block </p>
\[
a^*=x-b^*
\]
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample23(t *testing.T) {
	input := `Block \[a^*=x-b^*
\] equation`
	expected := `<p>Block </p>
\[a^*=x-b^*
\]
<p> equation</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample24(t *testing.T) {
	input := `Block \[
a^*=x-b^*\] equation`
	expected := `<p>Block </p>
\[
a^*=x-b^*\]
<p> equation</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample25(t *testing.T) {
	input := `$$
\begin{array} {lcl}
	L(p,w_i) &=& \dfrac{1}{N}\Sigma_{i=1}^N(\underbrace{f_r(x_2
	\rightarrow x_1
	\rightarrow x_0)G(x_1
	\longleftrightarrow x_2)f_r(x_3
	\rightarrow x_2
	\rightarrow x_1)}_{sample\, radiance\, evaluation\, in\, stage2}
	\\\\\\ &=&
	\prod_{i=3}^{k-1}(\underbrace{\dfrac{f_r(x_{i+1}
	\rightarrow x_i
	\rightarrow x_{i-1})G(x_i
	\longleftrightarrow x_{i-1})}{p_a(x_{i-1})}}_{stored\,in\,vertex\, during\,light\, path\, tracing\, in\, stage1})\dfrac{G(x_k
	\longleftrightarrow x_{k-1})L_e(x_k
	\rightarrow x_{k-1})}{p_a(x_{k-1})p_a(x_k)})
\end{array}
$$`

	expected := input
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample26(t *testing.T) {
	input := `\[
\begin{array} {lcl}
L(p,w_i) &=& \dfrac{1}{N}\Sigma_{i=1}^N(\underbrace{f_r(x_2
\rightarrow x_1
\rightarrow x_0)G(x_1
\longleftrightarrow x_2)f_r(x_3
\rightarrow x_2
\rightarrow x_1)}_{sample\, radiance\, evaluation\, in\, stage2}
\\\\\\ &=&
\prod_{i=3}^{k-1}(\underbrace{\dfrac{f_r(x_{i+1}
\rightarrow x_i
\rightarrow x_{i-1})G(x_i
\longleftrightarrow x_{i-1})}{p_a(x_{i-1})}}_{stored\,in\,vertex\, during\,light\, path\, tracing\, in\, stage1})\dfrac{G(x_k
\longleftrightarrow x_{k-1})L_e(x_k
\rightarrow x_{k-1})}{p_a(x_{k-1})p_a(x_k)})
\end{array}
\]`

	expected := input
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestRepeatedBlockNodesInOneParagraph(t *testing.T) {
	input := `Block $$x$$ equation $$y$$.`
	expected := `<p>Block </p>
$$x$$
<p> equation </p>
$$y$$
<p>.</p>`
	actual := Parse(t, input)
	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample27(t *testing.T) {
	input := `Block $$a^*=x-b^*$$ equation

Inline $a^*=x-b^*$ equation`
	expected := `<p>Block </p>
$$a^*=x-b^*$$
<p> equation</p>
<p>Inline $a^*=x-b^*$ equation</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample28(t *testing.T) {
	input := `Inline $a^*=x-b^*$ equation

Block $$a^*=x-b^*$$ equation`
	expected := `<p>Inline $a^*=x-b^*$ equation</p>
<p>Block </p>
$$a^*=x-b^*$$
<p> equation</p>`
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func BenchmarkWithAndWithoutPassthrough(b *testing.B) {
	const input = `
## Block
	
$$
a^*=x-b^*
$$

## Inline

Inline $a^*=x-b^*$ equation.`

	b.Run("without passthrough", func(b *testing.B) {
		md := goldmark.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with passthrough", func(b *testing.B) {
		md := buildTestParser()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if err := md.Convert([]byte(input), &buf); err != nil {
				b.Fatal(err)
			}
		}
	})
}
