package tag

import "github.com/vladvelici/spdx-go/spdx"

// ParseError represents both parsing and lexing errors
// It includes *spdx.Meta data (LineStart and LineEnd).
type ParseError struct {
	msg string
	*spdx.Meta
}

func (e *ParseError) Error() string {
	return e.msg
}

// Create a new *parseError with the given error message and *spdx.Meta
func parseError(msg string, m *spdx.Meta) *ParseError {
	return &ParseError{msg, m}
}
