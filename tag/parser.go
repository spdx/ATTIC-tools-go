package tag

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	OPEN_TAG  = "<text>"
	CLOSE_TAG = "</text>"
)

var (
	ErrNoCloseTag    = errors.New("Text tag opened but not closed. Missing a </text>?")
	ErrInvalidText   = errors.New("Some invalid formatted string found.")
	ErrInvalidPrefix = errors.New("No text is allowed between : and <text>.")
	ErrInvalidSuffix = errors.New("No text is allowed after close text tag (</text>).")
	ErrKeyNoValue    = errors.New("Key with no value returned by tokenizer")
)

// Returns the number of bytes that are (part of) unicode whitespace characters at the beginning of the given []byte.
func countSpaces(data []byte) int {
	width, start := 0, 0
	for ; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			return start
		}
	}
	return start
}

func tokenize() bufio.SplitFunc {
	hasKey := false
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		shifted := 0

		if !hasKey {
			a := true
			for a {
				a = false

				// clean whitespace
				spaces := countSpaces(data)

				// If all the data we have is white spaces, throw it away and read more.
				if spaces == len(data) {
					if atEOF {
						return 0, nil, nil
					}
					return shifted + spaces, nil, nil
				}

				// If there is other data as well, throw the spaces away and continue
				if spaces > 0 {
					shifted += spaces
					data = data[spaces:]
					a = true
				}

				// First character is # then the line is a comment
				if data[0] == '#' {
					endl := bytes.IndexByte(data, '\n')

					if endl < 0 {
						return shifted, nil, nil
					}

					// If the endline is the last character, throw everything away and read more
					if endl == len(data)-1 {
						if atEOF {
							return 0, nil, nil
						}
						return shifted + endl + 1, nil, nil
					}

					// Otherwise, if there is an endline and more data, throw the comment away and continue
					if endl >= 0 {
						a = true
						shifted += endl + 1
						data = data[endl+1:]
					}
				}
			}
		}

		if !hasKey {
			column := bytes.IndexByte(data, ':')
			if column < 0 {
				if atEOF {
					return 0, nil, ErrInvalidText
				}
				return shifted, nil, nil
			}
			hasKey = true
			return shifted + column + 1, data[:column], nil
		}

		startText := bytes.Index(data, []byte(OPEN_TAG))
		endl := bytes.IndexByte(data, '\n')

		if startText >= 0 && (endl < 0 || startText < endl) {

			if countSpaces(data[:startText]) != startText {
				return 0, nil, ErrInvalidPrefix
			}

			endText := bytes.Index(data, []byte(CLOSE_TAG))
			if endText < 0 {
				if atEOF {
					return 0, nil, ErrNoCloseTag
				}
				return shifted, nil, nil
			}

			valStart, valEnd := startText+len(OPEN_TAG), endText

			endlAfterEndTxt := bytes.IndexByte(data[endText:], '\n')

			var closeToEndl []byte

			if endlAfterEndTxt < 0 {
				if !atEOF {
					return 0, nil, nil
				}
				closeToEndl = data[endText+len(CLOSE_TAG):]
			} else if endlAfterEndTxt+endText < len(data) {
				closeToEndl = data[endText+len(CLOSE_TAG) : endlAfterEndTxt+endText]
			}

			if closeToEndl != nil && countSpaces(closeToEndl) != len(closeToEndl) {
				return 0, nil, ErrInvalidSuffix
			}

			hasKey = false
			return shifted + endText + len(CLOSE_TAG), data[valStart:valEnd], nil
		}

		if endl < 0 {
			if atEOF {
				hasKey = false
				return shifted + len(data), data[:], nil
			}
			return shifted, nil, nil
		}
		hasKey = false
		return shifted + endl + 1, data[:endl], nil
	}
}

type pair struct {
	Key, Value string
}

func parse(f io.Reader) (doc []pair, err error) {
	scanner := bufio.NewScanner(f)
	scanner.Split(tokenize())
	doc = make([]pair, 0, 15)
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		value := ""
		if scanner.Scan() {
			value = strings.TrimSpace(scanner.Text())
		}
		doc = append(doc, pair{key, value})
	}

	err = scanner.Err()
	return doc, err
}
