package tag

import "strings"
import "testing"

import "github.com/vladvelici/spdx-go/spdx"

func sameDoc(a, b []Pair) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestSameDocFunc(t *testing.T) {
	a := []Pair{
		{"one", "onev"},
		{"two", "twov"},
	}
	b := []Pair{
		{"one", "onev"},
		{"two", "twov"},
	}

	if !sameDoc(a, b) {
		t.Error("Slices are the same and not detected")
	}

	if sameDoc(a, []Pair{{"a", "c"}, {"two", "twov"}}) {
		t.Error("Slices are the same and detected as different")
	}
}

func TestEmptyDoc(t *testing.T) {
	r := strings.NewReader("")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}
}

func TestCommentsOnly(t *testing.T) {
	r := strings.NewReader("#this is a comment\n#this is another comment :)\n#whatever")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestOneCommentNoNewline(t *testing.T) {
	r := strings.NewReader("#this is a comment")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestOneCommentNewLine(t *testing.T) {
	r := strings.NewReader("#this is a comment\n")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestDoubleEndlineAfterComment(t *testing.T) {
	r := strings.NewReader("#property:value\n\n")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}
}

func TestCommentsAndEmptyLines(t *testing.T) {
	r := strings.NewReader("#this is a comment\n\n#this is another comment :)\n#whatever\n\n\n#anoterOne")

	doc, err := lexPair(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}
}

func TestValidProperty(t *testing.T) {
	r := strings.NewReader("someKey:someValue")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestValidPropertyNewLine(t *testing.T) {
	r := strings.NewReader("someKey:someValue\n")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestClearWhitespaces(t *testing.T) {
	r := strings.NewReader("someKey  : someValue\n")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestValidPropertyNewLineCR(t *testing.T) {
	r := strings.NewReader("someKey:someValue\r\n")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestValidPropertyText(t *testing.T) {
	r := strings.NewReader("someKey:<text>someValue</text>")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestValidPropertyTextMultiline(t *testing.T) {
	r := strings.NewReader("someKey:<text>\nsomeValue\n123\n\n4\n</text>")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue\n123\n\n4"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestValidPropertyTextNewLine(t *testing.T) {
	r := strings.NewReader("someKey:<text>someValue</text>\n")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestMoreInlineProperties(t *testing.T) {
	r := strings.NewReader("Property1:value1\nProperty2:value2\nProperty3:value3\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestInlinePropertiesAndComments(t *testing.T) {
	r := strings.NewReader("# comment\nProperty1:value1\nProperty2:value2\n# comment no two\nProperty3:value3\n#comm\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestInlinePropertiesCommentsAndNewlines(t *testing.T) {
	r := strings.NewReader("# comment\n\nProperty1:value1\n\n\nProperty2:value2\n# comment no two\nProperty3:value3\n#comm\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestMoreTextProperties(t *testing.T) {
	r := strings.NewReader("Property1:<text>value1</text>\nProperty2:<text>value2</text>\nProperty3:<text>value3</text>\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestMoreTextPropertiesAndComments(t *testing.T) {
	r := strings.NewReader("# this is a comment\nProperty1:<text>value1</text>\n#so is this\nProperty2:<text>value2</text>\nProperty3:<text>value3</text>\n#and this")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestMoreTextPropertiesCommentsAndNewlines(t *testing.T) {
	r := strings.NewReader("\n\n# this is a comment\n\nProperty1:<text>value1</text>\n#so is this\n\nProperty2:<text>value2</text>\nProperty3:<text>value3</text>\n#and this\n\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestMixedProperties(t *testing.T) {
	r := strings.NewReader("Property1:  <text>value1</text>\nProperty2:value2\nProperty3:<text>value3</text>\n")

	doc, err := lexPair(r)

	if len(doc) != 3 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

	properties := []Pair{
		{"Property1", "value1"},
		{"Property2", "value2"},
		{"Property3", "value3"},
	}

	if !sameDoc(properties, doc) {
		t.Errorf("Expected %s. Got %s", properties, doc)
	}
}

func TestInvalidTextValuePrefix(t *testing.T) {
	r := strings.NewReader("Property1: invalid <text>value1</text>\n")

	_, err := lexPair(r)

	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if e.Error() != MsgInvalidPrefix || (*e.Meta != spdx.Meta{1, 1}) {
		t.Errorf("Another error: %+v", err)
	}
}

func TestInvalidTextValueSuffix(t *testing.T) {
	r := strings.NewReader("Property1: <text>\n\nvalue1\n</text> invalid \n")

	_, err := lexPair(r)

	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if e.Error() != MsgInvalidSuffix || (*e.Meta != spdx.Meta{4, 4}) {
		t.Errorf("Another error: %s", err)
	}
}

func TestInvalidTextValueSuffixComment(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1</text># invalid \n")

	_, err := lexPair(r)

	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if e.Error() != MsgInvalidSuffix || (*e.Meta != spdx.Meta{1, 1}) {
		t.Errorf("Another error: %s", err)
	}
}

func TestInvalidTextValueSuffixProperty(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1</text>a:b\n")

	_, err := lexPair(r)
	t.Logf("Error: %s\n", err)

	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if e.Error() != MsgInvalidSuffix || (*e.Meta != spdx.Meta{1, 1}) {
		t.Errorf("Another error: %s", err)
	}
}

func TestInvalidUnclosedText(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1\n\n invalid \n")

	_, err := lexPair(r)

	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if err.Error() != MsgNoCloseTag || (*e.Meta != spdx.Meta{4, 4}) {
		t.Errorf("Another error: (%+v) %s", e.Meta, e)
	}
}

func TestInvalidProperty(t *testing.T) {
	r := strings.NewReader("Prop\nerty: value\n")

	doc, err := lexPair(r)

	t.Logf("doc len: %d", len(doc))

	e := err.(*ParseError)
	if e.msg != MsgInvalidText || e.Meta.LineStart != 1 {
		t.Errorf("Unexpected error: %+v", err)
	}

}

func TestCommentAsInlineValue(t *testing.T) {
	r := strings.NewReader("someKey:#someValue")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "#someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestCommentAsTextValue(t *testing.T) {
	r := strings.NewReader("someKey:<text>#someValue</text>")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "#someValue"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestCommentAsMultilineTextValue(t *testing.T) {
	r := strings.NewReader("someKey:<text>#c\n#someValue\nd\n#a</text>")

	doc, err := lexPair(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	p := Pair{"someKey", "#c\n#someValue\nd\n#a"}

	if len(doc) == 1 && doc[0] != p {
		t.Errorf("Expected %s. Got %s", p, doc[0])
	}
}

func TestSomeInvalidText(t *testing.T) {
	r := strings.NewReader("garbage")

	_, err := lexPair(r)
	if err == nil {
		t.Fail()
	}
	e := err.(*ParseError)
	if e.msg != MsgInvalidText || e.Meta.LineStart != 1 {
		t.Errorf("Unexpected error: %+v", err)
	}

}

func sameToken(a, b *Token) bool {
	return a == b || *a == *b || (*a.Meta == *b.Meta && a.Pair == b.Pair)
}

func sameTokens(a, b []Token) bool {
	if len(a) != len(b) {
		return false
	}

	for i, t := range a {
		if !sameToken(&t, &b[i]) {
			return false
		}
	}
	return true
}

func TestCommentToken(t *testing.T) {
	r := strings.NewReader("#comment1\n\n#comment2\n")

	tok, err := lexToken(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	expected := []Token{
		{TokenComment, Pair{"", "comment1"}, &spdx.Meta{1, 1}},
		{TokenComment, Pair{"", "comment2"}, &spdx.Meta{3, 3}},
	}

	if !sameTokens(tok, expected) {
		t.Errorf("Expected two comment tokens. Found: %+v", tok)
	}
}

func TestCommentAndProperty(t *testing.T) {
	r := strings.NewReader("#comment1\nprop:val\n")

	tok, err := lexToken(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	expected := []Token{
		{TokenComment, Pair{"", "comment1"}, &spdx.Meta{1, 1}},
		{TokenPair, Pair{"prop", "val"}, &spdx.Meta{2, 2}},
	}

	if !sameTokens(tok, expected) {
		t.Errorf("Wrong tokens. Found: %+v", tok)
	}
}

func TestLines(t *testing.T) {
	r := strings.NewReader("prop:<text>line1\nline2</text>\nprop:val\nprop2:val2")

	tok, err := lexToken(r)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	expected := []Token{
		{TokenPair, Pair{"prop", "line1\nline2"}, &spdx.Meta{1, 2}},
		{TokenPair, Pair{"prop", "val"}, &spdx.Meta{3, 3}},
		{TokenPair, Pair{"prop2", "val2"}, &spdx.Meta{4, 4}},
	}

	if !sameTokens(tok, expected) {
		t.Errorf("Wrong tokens. Found: %+v", tok)
	}
}

func TestPropertyWithNoValue(t *testing.T) {
	r := strings.NewReader("garbage:")
	doc, err := lexPair(r)
	t.Logf("Doc=%s, Err=%s", doc, err)
	p := Pair{"garbage", ""}
	if err != nil || doc == nil || len(doc) != 1 {
		t.FailNow()
	}
	if doc[0] != p {
		t.Fail()
	}
}

func TestAllDataWhitespaceAtEOF(t *testing.T) {
	lex := NewLexer(nil)
	lex.IgnoreComments = true
	f := lex.tokenizer()
	data := []byte("  \n")
	advance, token, err := f(data, true)
	if advance != 0 || token != nil || err != nil {
		t.Errorf("Fail with: advance=%d, data=%s, err=%s\n", advance, token, err)
	}
}

func TestCommentEndingInNewlineAtEOF(t *testing.T) {
	lex := NewLexer(nil)
	lex.IgnoreComments = true
	f := lex.tokenizer()
	data := []byte("#comment\n")
	advance, token, err := f(data, true)
	if advance != 0 || token != nil || err != nil {
		t.Errorf("Fail with: advance=%d, data=%s, err=%s\n", advance, token, err)
	}
}
