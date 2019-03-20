package tag

import "github.com/spdx/tools-go/spdx"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	openTag     = "<text>"
	closeTag    = "</text>"
	propertySep = ':'
)

// Types of tokens returned by tokenizer(). Only TokenComment and TokenPair are returned by Lexer.
const (
	TokenComment = iota // Token of type Comment
	tokenKey     = iota // Internally used by Lexer
	tokenValue   = iota // Internally used by Lexer
	TokenPair    = iota // Token of type Pair
)

// Represents a token returned by the Lexer
type Token struct {
	Type int
	Pair
	*spdx.Meta
}

// Better string representation for the token.
func (t *Token) String() string {
	if t.Type == TokenComment {
		return fmt.Sprintf("Comment{%s (%v)}", t.Pair.Value, t.Meta)
	}
	return fmt.Sprintf("Pair{%+v (%v)}", t.Pair, t.Meta)
}

// Create a new *Token with Type==TokenPair with the given key and value as the pair.
// Meta arguments:
// - if no meta arguments, returned *Token.Meta will be nil.
// - if 1 meta argument, returned *Token.Meta will be a new spdx.Meta pointer
//   with StartLine == EndLine == meta argument provided
// - if 2 or more meta arguments, returned *Token.Meta will be anew spdx.MEta pointer
//   with StartLine == first meta arugment and EndLine == second meta arguments.
//   The rest of arguments are ignored.
func PairTok(key, val string, meta ...int) *Token {
	var m *spdx.Meta
	if len(meta) >= 2 {
		m = &spdx.Meta{meta[0], meta[1]}
	} else if len(meta) == 1 {
		m = &spdx.Meta{meta[0], meta[0]}
	}
	return &Token{TokenPair, Pair{key, val}, m}
}

// Create a new *Token with Type==TokenComment and the given value as Pair.Value.
// See PairTok() for meta arguments usage.
func CommentTok(val string, meta ...int) *Token {
	var m *spdx.Meta
	if len(meta) >= 2 {
		m = &spdx.Meta{meta[0], meta[1]}
	} else if len(meta) == 1 {
		m = &spdx.Meta{meta[0], meta[0]}
	}
	return &Token{TokenComment, Pair{"", val}, m}
}

// Used to represent the value of a Token.
type Pair struct {
	Key, Value string
}

// Using this lexer interface so that we can easily make a fake lexer for testing.
type lexer interface {
	Lex() bool
	Token() *Token
	Err() error
	Line() int
}

// Tag format Lexer. Usage similar to bufio.Scanner
// The only Token types returned from the Lexer are TokenPair and TokenComment
// If IgnoreComments is set to true (default false), the lexer only returns Pair Tokens
type Lexer struct {
	r              io.Reader
	scanner        *bufio.Scanner
	line           int
	lineStart      int
	ttype          int
	token          *Token
	err            error
	IgnoreComments bool
	IgnoreMeta     bool
	CaseSensitive  bool
}

// Always use this function to create a new *Lexer. The reader given is used to form the tokens.
//
// The following settings are available:
// - IgnoreComments (default false).
//   Do not return Comment tokens. Ignore all lines that start with '#'.
// - IgnoreMeta (default false).
//   Do not assign spdx.Meta information (LineStart and LineEnd) to returned tokens.
// - CaseSensitive (default false)
//   If set to false, it tries to transform the case of Token.Pair.Key (of all non-comment tokens) to
//   match the one in the SPDX specification. E.g. transform "specversion" to "SpecVersion".
func NewLexer(r io.Reader) *Lexer {
	lexer := &Lexer{
		r:         r,
		scanner:   bufio.NewScanner(r),
		line:      1,
		lineStart: -1,
	}
	lexer.scanner.Split(lexer.tokenizer())
	return lexer
}

// Token gets the current token (must be called after Lex()).
// The returned *Token will be changed at the next call to Lex(). See Lex() for more.
func (l *Lexer) Token() *Token {
	return l.token
}

// Lex the next token. Returns true if there is a next token, false otherwise.
//
// There is only one token created by a Lexer and its properties are updated. Storing
// the pointer returned by Lexer.Token() itself will result in having its values changed
// at the next call of Lex().
//
// If there is an error while lexing, this method returns false and the
// error will be available by calling Err().
func (l *Lexer) Lex() bool {
	if !l.scanner.Scan() {
		l.err = l.scanner.Err()
		l.token = nil
		return false
	}

	if l.token == nil {
		l.token = new(Token)
	}

	if l.ttype == TokenComment {
		l.token.Type = TokenComment
		l.token.Pair.Key = ""
		l.token.Pair.Value = l.scanner.Text()
		if !l.IgnoreMeta {
			l.token.Meta = &spdx.Meta{l.line, l.line}
		}
		return true
	}

	// not comment, thus must be property
	l.token.Type = TokenPair

	l.token.Pair.Key = strings.TrimSpace(l.scanner.Text())
	if !l.CaseSensitive {
		ok, ci := IsValidPropertyInsensitive(l.token.Pair.Key)
		if ok && ci != l.token.Pair.Key {
			l.token.Pair.Key = ci
		}
	}

	l.token.Pair.Value = "" // empty string if no value found
	if l.scanner.Scan() {
		l.token.Pair.Value = strings.TrimSpace(l.scanner.Text())
	}

	if !l.IgnoreMeta {
		l.token.Meta = &spdx.Meta{l.line, l.line}
		// in case of multiline <text>:
		if l.lineStart > 0 {
			l.token.LineStart = l.lineStart
		}
	}

	return true
}

// Err gets the last error. If there is an error, it is either an I/O error returned initially by the associated io.Reader
// or an error of type *ParseError.
func (l *Lexer) Err() error {
	return l.err
}

// Line returns the line of last token (end line). This property is available even when IgnoreMeta is set to true.
// Use Token.Meta properties Token.LineStart and Token.LineEnd when those are available.
func (l *Lexer) Line() int {
	return l.line
}

// Returns the split function used by bufio.Scanner.
// Only used internally.
func (l *Lexer) tokenizer() bufio.SplitFunc {
	hasKey := false
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		l.lineStart = -1
		shifted := 0

		if !hasKey {
			a := true
			for a {
				a = false

				// clean whitespace
				spaces, newlines := countSpacesNl(data)
				l.line += newlines

				// If all the data we have is white spaces, throw it away and read more.
				if spaces == len(data) {
					if atEOF {
						return 0, nil, nil
					}
					return shifted + spaces, nil, nil
				}

				// If there is other data as well, throw the spaces away and continue
				if spaces > 0 {
					shifted += spaces
					data = data[spaces:]
					a = true
				}

				// First character is # then the line is a comment
				if data[0] == '#' {
					endl := bytes.IndexByte(data, '\n')

					// If there is no endline, read more
					if endl < 0 {
						return shifted, nil, nil
					}

					// If we want the comments as tokens
					if !l.IgnoreComments {
						l.ttype = TokenComment
						return shifted + endl, data[1:endl], nil
					}

					// If the endline is the last character, throw everything away and read more
					if endl == len(data)-1 {
						if atEOF {
							return 0, nil, nil
						}
						return shifted + endl + 1, nil, nil
					}

					// Otherwise, if there is an endline and more data, throw the comment away and continue
					if endl >= 0 {
						l.line++
						a = true
						shifted += endl + 1
						data = data[endl+1:]
					}
				}
			}
		}

		// Find a property (key)
		if !hasKey {
			column := bytes.IndexByte(data, propertySep)
			endl := bytes.IndexByte(data, '\n')

			if endl >= 0 && endl < column {
				return 0, nil, spdx.NewParseError(MsgInvalidText, &spdx.Meta{l.line, l.line})
			}

			if column < 0 {
				if atEOF {
					return 0, nil, spdx.NewParseError(MsgInvalidText, &spdx.Meta{l.line, l.line})
				}
				return shifted, nil, nil
			}
			l.ttype = tokenKey
			hasKey = true
			return shifted + column + 1, data[:column], nil
		}

		// we have a property, find a value
		startText := bytes.Index(data, []byte(openTag))
		endl := bytes.IndexByte(data, '\n')

		// if <text> found before newline
		if startText >= 0 && (endl < 0 || startText < endl) {

			l.lineStart = l.line // lineStart is at the start of property
			if countSpaces(data[:startText]) != startText {
				return 0, nil, spdx.NewParseError(MsgInvalidPrefix, &spdx.Meta{l.line, l.line})
			}

			endText := bytes.Index(data, []byte(closeTag))
			if endText < 0 {
				if atEOF {
					l.line += bytes.Count(data, []byte{'\n'})
					return 0, nil, spdx.NewParseError(MsgNoCloseTag, &spdx.Meta{l.line, l.line})
				}
				return shifted, nil, nil
			}

			valStart, valEnd := startText+len(openTag), endText
			l.line += bytes.Count(data[:valEnd], []byte{'\n'}) // count lines in the value

			// check if there's anything else than spaces after </text>
			endlAfterEndTxt := bytes.IndexByte(data[endText:], '\n')

			var closeToEndl []byte

			if endlAfterEndTxt < 0 {
				if !atEOF {
					return 0, nil, nil
				}
				closeToEndl = data[endText+len(closeTag):]
			} else if endlAfterEndTxt+endText < len(data) {
				closeToEndl = data[endText+len(closeTag) : endlAfterEndTxt+endText]
			}

			if closeToEndl != nil && countSpaces(closeToEndl) != len(closeToEndl) {
				return 0, nil, spdx.NewParseError(MsgInvalidSuffix, &spdx.Meta{l.line, l.line})
			}

			hasKey = false
			l.ttype = tokenValue
			return shifted + endText + len(closeTag), data[valStart:valEnd], nil
		}

		// if no <text> we have an inline property
		if endl < 0 {
			if atEOF {
				l.ttype = tokenValue
				hasKey = false
				return shifted + len(data), data, nil
			}
			return shifted, nil, nil
		}

		l.ttype = tokenValue
		hasKey = false
		return shifted + endl, data[:endl], nil
	}
}

// Lex all the pairs.
func lexPair(f io.Reader) ([]Pair, error) {
	p := make([]Pair, 0)
	lex := NewLexer(f)
	lex.IgnoreComments = true

	for lex.Lex() {
		p = append(p, lex.Token().Pair)
	}

	if lex.Err() != nil {
		return nil, lex.Err()
	}

	return p, nil
}

// Lex all the tokens.
func lexToken(f io.Reader) ([]*Token, error) {
	p := make([]*Token, 0)
	lex := NewLexer(f)
	lex.IgnoreComments = false

	for lex.Lex() {
		tok := *lex.Token()
		p = append(p, &tok)
	}

	if lex.Err() != nil {
		return nil, lex.Err()
	}

	return p, nil
}

// Returns the number of bytes that are (part of) unicode whitespace characters
// at the beginning of the given []byte and the number of lines these include.
func countSpacesNl(data []byte) (spaces int, lines int) {
	width, start := 0, 0
	for ; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if r == '\n' {
			lines++
		}
		if !unicode.IsSpace(r) {
			return start, lines
		}
	}
	return start, lines
}

// Returns the number of bytes that are (part of) unicode whitespace characters
// at the beginning of the given []byte.
func countSpaces(data []byte) int {
	width, start := 0, 0
	for ; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			return start
		}
	}
	return start
}
