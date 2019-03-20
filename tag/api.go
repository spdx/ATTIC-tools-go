package tag

import "github.com/spdx/tools-go/spdx"

import (
	"io"
)

var (
	// Build() settings for Lexer.IgnoreMeta
	noMeta = false

	// Build() settings for Lexer.CaseSensitive
	caseSensitive = false
)

// Set the Lexer.IgnoreMeta used by Build() function. Default is false.
func IgnoreMeta(meta bool) { noMeta = meta }

// GetIgnoreMeta gets the current option used by Lexer.IgnoreMeta used by Build().
func GetIgnoreMeta() bool { return noMeta }

// Set the Lexer.CaseSensitive option used by Build() function. Default is false.
func CaseSensitive(ci bool) { caseSensitive = ci }

// GetCaseSensitive gets the current option for Lexer.IgnoreCase used by Build().
func GetCaseSensitive() bool { return caseSensitive }

// Lex a io.Reader and Parse it to a *spdx.Document. If there is an error, it is
// of type *ParseError.
func Build(f io.Reader) (*spdx.Document, error) {
	lexer := NewLexer(f)
	lexer.IgnoreComments = true
	lexer.IgnoreMeta = noMeta
	lexer.CaseSensitive = caseSensitive
	return Parse(lexer)
}

// Write a *spdx.Document to the given io.Writer
func Write(f io.Writer, doc *spdx.Document) error {
	p := NewFormatter(f)
	return p.Document(doc)
}
