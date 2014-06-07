package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"errors"
	"io"
)

func Parse(f io.Reader) (*spdx.Document, error) {
	pairs, err := parse(f)
	if err != nil {
		return nil, err
	}
	return build(pairs)
}

func Write(f io.Writer, doc *spdx.Document) error {
	return writeDocument(f, doc)
}
