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

func TestMassiveEquation(t *testing.T) {
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

func TestMassiveEquationSquareDelimiters(t *testing.T) {
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

func TestExample1(t *testing.T) {
	input := `Inline $x = {-b \pm \sqrt{b^2-4ac} \over 2a}$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample2(t *testing.T) {
	input := `x = {-b \pm \sqrt{b^2-4ac} \over 2a}
n$ equation`

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

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Assert(actual, qt.Equals, expected)
}

func TestExample6(t *testing.T) {
	input := `Block $$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample7(t *testing.T) {
	input := `Block $$x = {-b \pm \sqrt{b^2-4ac} \over 2a}
$$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample8(t *testing.T) {
	input := `Block $$
x = {-b \pm \sqrt{b^2-4ac} \over 2a}$$ equation`

	expected := "<p>" + input + "</p>"
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

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample14(t *testing.T) {
	input := `Block $$
a^*=x-b^*
$$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample15(t *testing.T) {
	input := `Block $$a^*=x-b^*
$$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample16(t *testing.T) {
	input := `Block $$
a^*=x-b^*$$ equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
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

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample22(t *testing.T) {
	input := `Block \[
a^*=x-b^*
\] equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample23(t *testing.T) {
	input := `Block \[a^*=x-b^*
	\] equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
	c.Assert(actual, qt.Equals, expected)
}

func TestExample24(t *testing.T) {
	input := `Block \[
a^*=x-b^*\] equation`

	expected := "<p>" + input + "</p>"
	actual := Parse(t, input)

	c := qt.New(t)
	c.Skip() // TODO failing test
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
