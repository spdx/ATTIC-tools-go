package spdx

// File Types
const (
	FT_BINARY  = iota
	FT_SOURCE  = iota
	FT_ARCHIVE = iota
	FT_OTHER   = iota
)

// supported specification versions
var SpecVersions = [][2]int{{1, 2}}

const (
	NOASSERTION = "NOASSERTION"
	NONE        = "NONE"
)

type Value interface {
	V() string
	M() *Meta
}

type ValueStr struct {
	Val  string
	Meta *Meta
}

func Str(v string, m *Meta) ValueStr     { return ValueStr{v, m} }
func (v ValueStr) V() string             { return v.Val }
func (v ValueStr) M() *Meta              { return v.Meta }
func (v ValueStr) Equal(w ValueStr) bool { return v.Val == w.Val }

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

type ValueBool struct {
	Val  bool
	Meta *Meta
}

func Bool(v bool, m *Meta) ValueBool { return ValueBool{v, m} }
func (v ValueBool) V() string {
	if v.Val {
		return "true"
	}
	return "false"
}
func (v ValueBool) M() *Meta { return v.Meta }

type Meta struct {
	LineStart, LineEnd int
}

func NewMetaL(line int) *Meta {
	return &Meta{line, line}
}

func NewMeta(start, end int) *Meta {
	return &Meta{start, end}
}
