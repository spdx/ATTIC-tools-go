package tag

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode"
	"unicode/utf8"
)

const (
	OPEN_TAG  = "<text>"
	CLOSE_TAG = "</text>"
)

func skipSpaces(data []byte) (width int) {
	start := 0
	for width = 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			return
		}
	}
	return
}

func tokenize() bufio.SplitFunc {
	hasKey := false
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// remove comments
		if data[0] == '#' {
			if endl := bytes.IndexByte(data, '\n'); endl < 0 || endl == len(data)-1 {
				return 0, nil, nil
			} else {
				return endl + 1, nil, nil
			}
		}

		// skip spaces and blank lines
		if spaces := skipSpaces(data); spaces > 0 {
			return spaces + 1, nil, nil
		}

		if !hasKey {
			column := bytes.IndexByte(data, ':')
			if column < 0 {
				if atEOF {
					return 0, nil, errors.New("No key but some text found")
				}
				return 0, nil, nil
			}
			hasKey = true
			return column + 1, data[:column], nil
		}

		if startText := bytes.IndexAny(data, OPEN_TAG); startText > 0 {
			endText := bytes.IndexAny(data, CLOSE_TAG)
			if endText < 0 {
				if atEOF {
					return 0, nil, errors.New("<text> tag not closed.")
				}
				return 0, nil, nil
			}
			valStart, valEnd := startText+len(OPEN_TAG), endText-1
			hasKey = false
			return endText + len(CLOSE_TAG), data[valStart:valEnd], nil
		}

		endl := bytes.IndexByte(data, '\n')
		if endl < 0 && !atEOF {
			return 0, nil, nil
		}
		if atEOF {
			return len(data) + 1, data, nil
		}
		return endl + 1, data[:endl], nil
	}
}

type Pair struct {
	Key, Value string
}

func Parse(f io.Reader) (doc []Pair, err error) {
	scanner := bufio.NewScanner(f)
	scanner.Split(tokenize())

	doc = make([]Pair, 0, 15)

	for scanner.Scan() {
		if scanner.Bytes() == nil {
			continue
		}
		key := scanner.Text()
		if !scanner.Scan() {
			return nil, errors.New("Key with no value returned by tokenizer")
		}
		value := scanner.Text()
		doc = append(doc, Pair{key, value})
	}

	err = scanner.Err()
	return doc, err
}
