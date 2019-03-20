package rdf

import (
	"github.com/deltamobile/goraptor"
	"strings"
)

// Constants representing RDF formats supported by raptor.
//
// One of the accepted formats not in this constants is "rdf". When parsing,
// "rdf" means using raptor's "guess" parser. When writing, it means using
// Fmt_rdfxmlAbbrev.
const (
	Fmt_ntriples     = "ntriples"      // for N-Triples
	Fmt_turtle       = "turtle"        // for Turtle Terse RDF Triple Language
	Fmt_rdfxmlXmp    = "rdfxml-xmp"    // for RDF/XML (XMP Profile)
	Fmt_rdfxmlAbbrev = "rdfxml-abbrev" // for RDF/XML (Abbreviated)
	Fmt_rdfxml       = "rdfxml"        // for RDF/XML
	Fmt_rss          = "rss-1.0"       // for RSS 1.0
	Fmt_atom         = "atom"          // for Atom 1.0
	Fmt_dot          = "dot"           // for GraphViz DOT format
	Fmt_jsonTriples  = "json-triples"  // for RDF/JSON Triples
	Fmt_json         = "json"          // for RDF/JSON Resource-Centric
	Fmt_html         = "html"          // for HTML Table
	Fmt_nquads       = "nquads"        // for N-Quads
)

// Useful RDF URIs
const (
	baseUri    = "http://spdx.org/rdf/terms#"
	licenceUri = "http://spdx.org/licenses/"
)

// Common RDF prefixes used in SPDX RDF Representations.
var rdfPrefixes = map[string]string{
	"ns:":   "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"doap:": "http://usefulinc.com/ns/doap#",
	"rdfs:": "http://www.w3.org/2000/01/rdf-schema#",
	"":      baseUri,
}

// Useful helper pair struct.
type pair struct {
	key, val string
}

// FormatOk checks if `fmt` is one of the raptor supported formats (has the value of one
// of the Fmt_* constants). The special "rdf" value is considered invalid by
// this function.
func FormatOk(fmt string) bool {
	fmts := []string{
		Fmt_ntriples,
		Fmt_turtle,
		Fmt_rdfxmlXmp,
		Fmt_rdfxmlAbbrev,
		Fmt_rdfxml,
		Fmt_rss,
		Fmt_atom,
		Fmt_dot,
		Fmt_jsonTriples,
		Fmt_json,
		Fmt_html,
		Fmt_nquads,
	}
	for _, f := range fmts {
		if fmt == f {
			return true
		}
	}
	return false
}

// Expands the prefixes "ns:", "doap:" and "rdfs:" to their full URIs.
// If there is no ":" or there is another prefix, it expands to baseUri.
func prefix(k string) *goraptor.Uri {
	var pref string
	rest := k
	if i := strings.Index(k, ":"); i >= 0 {
		pref = k[:i+1]
		rest = k[i+1:]
	}
	if long, ok := rdfPrefixes[pref]; ok {
		pref = long
	}
	uri := goraptor.Uri(pref + rest)
	return &uri
}

// Change the RDF prefixes to their short forms.
func shortPrefix(t goraptor.Term) string {
	str := termStr(t)
	for short, long := range rdfPrefixes {
		if strings.HasPrefix(str, long) {
			return strings.Replace(str, long, short, 1)
		}
	}
	return str
}

// goraptor.Term to string. Returns empty string if the term given is not one of
// the following types: *goraptor.Uri, *goraptor.Blank or *goraptor.Literal.
func termStr(term goraptor.Term) string {
	switch t := term.(type) {
	case *goraptor.Uri:
		return string(*t)
	case *goraptor.Blank:
		return string(*t)
	case *goraptor.Literal:
		return t.Value
	default:
		return ""
	}
}

// Create *goraptor.Uri from string
func uri(uri string) *goraptor.Uri {
	return (*goraptor.Uri)(&uri)
}

// Create *goraptor.Literal from string
func literal(lit string) *goraptor.Literal {
	return &goraptor.Literal{Value: lit}
}

// Create *goraptor.Blank from string
func blank(b string) *goraptor.Blank {
	return (*goraptor.Blank)(&b)
}
