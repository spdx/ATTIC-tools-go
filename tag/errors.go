package tag

import "github.com/vladvelici/spdx-go/spdx"

// Error messages used by the parser
var (
	MsgNoClosedParen             = "No closed parentheses at the end."
	MsgInvalidVerifCodeExcludes  = "VerificationCode: Invalid Excludes format"
	MsgInvalidChecksum           = "Invalid Package Checksum format."
	MsgConjunctionAndDisjunction = "Licence sets can only have either disjunction or conjunction, not both. (AND or OR, not both)"
	MsgEmptyLicence              = "Empty licence"
	MsgAlreadyDefined            = "Property already defined"
)

// Error messages used by the lexer
var (
	MsgNoCloseTag    = "Text tag opened but not closed. Missing a </text>?"
	MsgInvalidText   = "Some invalid formatted string found."
	MsgInvalidPrefix = "No text is allowed between : and <text>."
	MsgInvalidSuffix = "No text is allowed after close text tag (</text>)."
)

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
