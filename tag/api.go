package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"io"
)

// Lex a io.Reader and Parse it to a *spdx.Document
func Build(f io.Reader) (*spdx.Document, error) {
	lexer := NewLexer(f)
	lexer.IgnoreComments = true
	return Parse(lexer)
}

// Write a *spdx.Document to the given io.Writer
func Write(f io.Writer, doc *spdx.Document) error {
	p := NewFormatter(f)
	return p.Document(doc)
}
