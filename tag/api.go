package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"io"
)

var (
	// Build() settings for Lexer.IgnoreMeta
	noMeta = false

	// Build() settings for Lexer.CaseSensitive
	noCase = false
)

// Set the Lexer.IgnoreMeta used by Build() function. Default is false.
func IgnoreMeta(meta bool) { noMeta = meta }

// Get the current option used by Lexer.IgnoreMeta used by Build().
func GetIgnoreMeta() bool { return noMeta }

// Set the Lexer.IgnoreCase option used by Build() function. Default is false.
func IgnoreCase(ci bool) { noCase = ci }

// Get the current option for Lexer.IgnoreCase used by Build().
func GetIgnoreCase() bool { return noCase }

// Lex a io.Reader and Parse it to a *spdx.Document. If there is an error, it is of type *ParseError.
func Build(f io.Reader) (*spdx.Document, error) {
	lexer := NewLexer(f)
	lexer.IgnoreComments = true
	lexer.IgnoreMeta = noMeta
	lexer.IgnoreCase = noCase
	return Parse(lexer)
}

// Write a *spdx.Document to the given io.Writer
func Write(f io.Writer, doc *spdx.Document) error {
	p := NewFormatter(f)
	return p.Document(doc)
}
