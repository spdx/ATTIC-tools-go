package spdx

import (
	"regexp"
	"strings"
	"time"
)

// File Types
const (
	FT_BINARY      = "BINARY"
	FT_SOURCE      = "SOURCE"
	FT_ARCHIVE     = "ARCHIVE"
	FT_OTHER       = "OTHER"
	FT_APPLICATION = "APPLICATION"
	FT_AUDIO       = "AUDIO"
	FT_IMAGE       = "IMAGE"
	FT_TEXT        = "TEXT"
	FT_VIDEO       = "VIDEO"
)

// supported specification versions
var SpecVersions = [][2]int{{1, 2}}

// Regex for the Creator format: `What: Who (email)`
var CreatorRegex = regexp.MustCompile("^([^:]*):([^\\(]*)(\\((.*)\\))?$")

// SPDX value constants
const (
	NOASSERTION = "NOASSERTION"
	NONE        = "NONE"
)

// Interface to be used for SPDX Elements.
// Implemented by Value(Str|Bool|Date|Creator)
type Value interface {
	V() string
	M() *Meta
}

// Store a string with relevant metadata
type ValueStr struct {
	Val  string
	Meta *Meta
}

// Create a new ValueStr form v and m.
func Str(v string, m *Meta) ValueStr     { return ValueStr{v, m} }
func (v ValueStr) V() string             { return v.Val }
func (v ValueStr) M() *Meta              { return v.Meta }
func (v ValueStr) Equal(w ValueStr) bool { return v.Val == w.Val }

// Store a boolean value with relevant metadata
type ValueBool struct {
	Val  bool
	Meta *Meta
}

func Bool(v bool, m *Meta) ValueBool { return ValueBool{v, m} }
func (v ValueBool) M() *Meta         { return v.Meta }
func (v ValueBool) V() string {
	if v.Val {
		return "true"
	}
	return "false"
}

// Store data similar to document creator or package spplier or originator.
// If data stored using SetValue() is of the form `what: name (email)`, where
// `(email)` is optional, the `what`, `name` and `email` fields get populated.
type ValueCreator struct {
	val   string
	what  string
	name  string
	email string
	*Meta
}

// V gets the original value of this ValueCreator
func (c ValueCreator) V() string { return c.val }

// M gets the metadata associated with this ValueCreator
func (c ValueCreator) M() *Meta { return c.Meta }

// What gets the `what` part from the format `what: name (email)`.
func (c ValueCreator) What() string { return c.what }

// Name gets the `name` part from the format `what: name (email)`
func (c ValueCreator) Name() string { return c.name }

// Email gets the `email` part from the format `what: name (email)`
func (c ValueCreator) Email() string { return c.email }

// Set the value of this ValueCreator. It parses the format `what: name (email)`
// and populates the relevant fields.
func (c *ValueCreator) SetValue(v string) {
	c.val = v
	match := CreatorRegex.FindStringSubmatch(v)
	if len(match) == 5 {
		c.what = strings.TrimSpace(match[1])
		c.name = strings.TrimSpace(match[2])
		c.email = strings.TrimSpace(match[4])
	}
}

// Create and populate a new ValueCreator.
func NewValueCreator(val string, m *Meta) ValueCreator {
	vc := ValueCreator{Meta: m}
	(&vc).SetValue(val)
	return vc
}

// Store Dates of format YYYY-MM-DDThh:mm:ssZ.
// If the time is in the correct format, it is available parsed into a
// *time.Time by calling Time().
type ValueDate struct {
	val  string
	time *time.Time
	*Meta
}

// V gets the original value of this ValueDate.
func (d ValueDate) V() string { return d.val }

// M gets the metadata of this ValueDate.
func (d ValueDate) M() *Meta { return d.Meta }

// Time gets the *time.Time pointer parsed form the value.
func (d ValueDate) Time() *time.Time { return d.time }

// Set the value of this ValueDate and parse the date format.
func (d *ValueDate) SetValue(v string) {
	d.val = v
	t, err := time.Parse(time.RFC3339, v)
	if err == nil {
		d.time = &t
	}
}

// Create and populate a new ValueDate.
func NewValueDate(val string, m *Meta) ValueDate {
	vd := ValueDate{Meta: m}
	(&vd).SetValue(val)
	return vd
}

// Store metadata about SPDX Elements
type Meta struct {
	LineStart, LineEnd int
}

// Create a new Meta with both lineStart and lineEnd set to line.
func NewMetaL(line int) *Meta {
	return &Meta{line, line}
}

// Create a new Meta with the given start and end lines.
func NewMeta(start, end int) *Meta {
	return &Meta{start, end}
}

// strings.Join for ValueStr type.
func Join(a []ValueStr, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return a[0].Val
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i].Val)
	}

	b := make([]byte, n)
	bp := copy(b, a[0].Val)
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s.Val)
	}
	return string(b)
}

// ParseError represents both parsing and lexing errors
// It includes *spdx.Meta data (LineStart and LineEnd).
type ParseError struct {
	msg string
	*Meta
}

// Error returns the error message.
func (e *ParseError) Error() string {
	return e.msg
}

// Create a new *ParseError with the given error message and *spdx.Meta
func NewParseError(msg string, m *Meta) *ParseError {
	return &ParseError{msg, m}
}
