# Hugo Goldmark Extensions

[![Tests on Linux, MacOS and Windows](https://github.com/gohugoio/hugo-goldmark-extensions/workflows/Test/badge.svg)](https://github.com/gohugoio/hugo-goldmark-extensions/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/gohugoio/hugo-goldmark-extensions)](https://goreportcard.com/report/github.com/gohugoio/hugo-goldmark-extensions)
[![GoDoc](https://godoc.org/github.com/gohugoio/hugo-goldmark-extensions?status.svg)](https://godoc.org/github.com/gohugoio/hugo-goldmark-extensions)

This repository houses a collection of [Goldmark] extensions created by the [Hugo] community, focusing on expanding Hugo's markdown functionality.

[CommonMark]: https://spec.commonmark.org/0.30/
[Goldmark]: https://github.com/yuin/goldmark/
[Hugo]: https://gohugo.io/
[LaTeX]: https://www.latex-project.org/about/
[KaTeX]: https://katex.org/
[MathJax]: https://www.mathjax.org/

## Passthrough extension

Use this extension to preserve raw markdown within delimited snippets of text. This was initially developed to support [LaTeX] mixed with markdown, specifically mathematical expressions and equations.

For example, to preserve raw markdown for inline snippets delimited by the `$` character:

Markdown|Default rendering|Passthrough rendering
:--|:--|:--
`a $_text_$ snippet`|`a $<em>text</em>$ snippet`|`a $_text_$ snippet`

In the markdown example above, the underscores surrounding the word "text" signify emphasis. The markdown renderer wraps the word within `em` tags as required by the [CommonMark] specification. In comparison, the passthrough extension preserves the text within and including the delimiters.

Why is this important? Consider this example of a mathematical equation written in LaTeX:

Markdown|Default rendering|Passthrough rendering
:--|:--|:--
`$a^*=x-b^*$`|`$a^<em>=x-b^</em>$`|`$a^*=x-b^*$`

Without this extension, LaTeX parsers such as [KaTeX] and [MathJax] will render this:

\$a^<em>=x-b^</em>\$

Instead of this:

$a^\*=x-b^\*$

### Delimiters

There are two types of delimiters, inline and block:

- Text within and including inline delimiters appears inline with the surrounding text.
- Text within and including block delimiters appears between adjacent block elements. When working with LaTeX this is known as "display" mode.

As shown below, you must specify zero or more inline and block delimiter pairs&mdash; there are no default values.

### Usage

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/gohugoio/hugo-goldmark-extensions/passthrough"
	"github.com/yuin/goldmark"
)

func main() {
	inlineDelimiters := []passthrough.Delimiters{
		{Open: "\\(", Close: "\\)"},
	}
	blockDelimiters := []passthrough.Delimiters{
		{Open: "\\[", Close: "\\]"},
		{Open: "$$", Close: "$$"},
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			passthrough.NewPassthroughWithDelimiters(
				inlineDelimiters,
				blockDelimiters,
			)),
	)

	input := `
block \[a^*=x-b^*\] snippet

inline \(a^*=x-b^*\) snippet
`

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
```
