package rdf

import (
	"github.com/deltamobile/goraptor"
	"strings"
)

const (
	baseUri    = "http://spdx.org/rdf/terms#"
	licenceUri = "http://spdx.org/licenses/"
)

type pair struct {
	key, val string
}

func prefix(k string) *goraptor.Uri {
	var pref string
	rest := k
	if i := strings.Index(k, ":"); i >= 0 {
		pref = k[:i]
		rest = k[i+1:]
	}

	switch pref {
	default:
		pref = baseUri
	case "ns":
		pref = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	case "doap":
		pref = "http://usefulinc.com/ns/doap#"
	case "rdfs":
		pref = "http://www.w3.org/2000/01/rdf-schema#"
	}

	uri := goraptor.Uri(pref + rest)
	return &uri
}

func uri(uri string) *goraptor.Uri {
	return (*goraptor.Uri)(&uri)
}

func literal(lit string) *goraptor.Literal {
	return &goraptor.Literal{Value: lit}
}

func blank(b string) *goraptor.Blank {
	return (*goraptor.Blank)(&b)
}
