package tag

import "github.com/vladvelici/spdx-go/spdx"

// Parse error represents both parsing and lexing errors
type ParseError struct {
	msg string
	*spdx.Meta
}

func (e *ParseError) Error() string {
	return e.msg
}

func parseError(msg string, m *spdx.Meta) *ParseError {
	return &ParseError{msg, m}
}
