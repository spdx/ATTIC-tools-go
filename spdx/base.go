package spdx

import (
	"regexp"
	"strings"
	"time"
)

// File Types
const (
	FT_BINARY  = iota
	FT_SOURCE  = iota
	FT_ARCHIVE = iota
	FT_OTHER   = iota
)

// supported specification versions
var SpecVersions = [][2]int{{1, 2}}

var CreatorRegex = regexp.MustCompile("^([^:]*):([^\\(]*)(\\((.*)\\))?$")

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
// If data stored using SetValue() is of the form `what: name (email)`, where `(email)` is optional,
// the `what`, `name` and `email` fields get populated.
type ValueCreator struct {
	val   string
	what  string
	name  string
	email string
	*Meta
}

func (c ValueCreator) V() string     { return c.val }
func (c ValueCreator) M() *Meta      { return c.Meta }
func (c ValueCreator) What() string  { return c.what }
func (c ValueCreator) Name() string  { return c.name }
func (c ValueCreator) Email() string { return c.email }
func (c *ValueCreator) SetValue(v string) {
	c.val = v
	match := CreatorRegex.FindStringSubmatch(v)
	if len(match) == 5 {
		c.what = strings.TrimSpace(match[1])
		c.name = strings.TrimSpace(match[2])
		c.email = strings.TrimSpace(match[4])
	}
}
func NewValueCreator(val string, m *Meta) ValueCreator {
	vc := ValueCreator{Meta: m}
	(&vc).SetValue(val)
	return vc
}

// Store Dates of format YYYY-MM-DDThh:mm:ssZ.
// If the time is in the correct format, it is available parsed into a *time.Time by calling Time().
type ValueDate struct {
	val  string
	time *time.Time
	*Meta
}

func (d ValueDate) V() string        { return d.val }
func (d ValueDate) M() *Meta         { return d.Meta }
func (d ValueDate) Time() *time.Time { return d.time }
func (d *ValueDate) SetValue(v string) {
	d.val = v
	t, err := time.Parse(time.RFC3339, v)
	if err == nil {
		d.time = &t
	}
}
func NewValueDate(val string, m *Meta) ValueDate {
	vd := ValueDate{Meta: m}
	(&vd).SetValue(val)
	return vd
}

// Store metadata about SPDX Elements
type Meta struct {
	LineStart, LineEnd int
}

func NewMetaL(line int) *Meta {
	return &Meta{line, line}
}

func NewMeta(start, end int) *Meta {
	return &Meta{start, end}
}

// strings.Join for ValueStr type
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
