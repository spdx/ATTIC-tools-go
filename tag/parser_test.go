package tag

import "strings"
import "testing"

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

	doc, err := Parse(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}
}

func TestCommentsOnly(t *testing.T) {
	r := strings.NewReader("#this is a comment\n#this is another comment :)\n#whatever")

	doc, err := Parse(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestOneCommentNoNewline(t *testing.T) {
	r := strings.NewReader("#this is a comment")

	doc, err := Parse(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestOneCommentNewLine(t *testing.T) {
	r := strings.NewReader("#this is a comment\n")

	doc, err := Parse(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}

}

func TestCommentsAndEmptyLines(t *testing.T) {
	r := strings.NewReader("#this is a comment\n\n#this is another comment :)\n#whatever\n\n\n#anoterOne\n\n")

	doc, err := Parse(r)

	if len(doc) != 0 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
	}
}

func TestValidProperty(t *testing.T) {
	r := strings.NewReader("someKey:someValue")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestValidPropertyNewLine(t *testing.T) {
	r := strings.NewReader("someKey:someValue\n")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestClearWhitespaces(t *testing.T) {
	r := strings.NewReader("someKey  : someValue\n")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestValidPropertyNewLineCR(t *testing.T) {
	r := strings.NewReader("someKey:someValue\r\n")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestValidPropertyText(t *testing.T) {
	r := strings.NewReader("someKey:<text>someValue</text>")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestValidPropertyTextMultiline(t *testing.T) {
	r := strings.NewReader("someKey:<text>\nsomeValue\n123\n\n4\n</text>")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue\n123\n\n4"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestValidPropertyTextNewLine(t *testing.T) {
	r := strings.NewReader("someKey:<text>someValue</text>\n")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestMoreInlineProperties(t *testing.T) {
	r := strings.NewReader("Property1:value1\nProperty2:value2\nProperty3:value3\n")

	doc, err := Parse(r)

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

	doc, err := Parse(r)

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

	doc, err := Parse(r)

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

	doc, err := Parse(r)

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

	doc, err := Parse(r)

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

	doc, err := Parse(r)

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
	r := strings.NewReader("Property1:<text>value1</text>\nProperty2:value2\nProperty3:<text>value3</text>\n")

	doc, err := Parse(r)

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

	_, err := Parse(r)

	if err == nil {
		t.Fail()
	}
}

func TestInvalidTextValueSuffix(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1</text> invalid \n")

	_, err := Parse(r)

	if err == nil {
		t.Fail()
	}
}

func TestInvalidTextValueSuffixComment(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1</text># invalid \n")

	_, err := Parse(r)

	if err == nil {
		t.Fail()
	}
}

func TestInvalidTextValueSuffixProperty(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1</text>a:b\n")

	_, err := Parse(r)

	if err == nil {
		t.Fail()
	}
}

func TestInvalidUnclosedText(t *testing.T) {
	r := strings.NewReader("Property1: <text>value1\n\n invalid \n")

	_, err := Parse(r)

	if err == nil {
		t.Fail()
	}
}

func TestCommentAsInlineValue(t *testing.T) {
	r := strings.NewReader("someKey:#someValue")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "#someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestCommentAsTextValue(t *testing.T) {
	r := strings.NewReader("someKey:<text>#someValue</text>")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "#someValue"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}

func TestCommentAsMultilineTextValue(t *testing.T) {
	r := strings.NewReader("someKey:<text>#c\n#someValue\nd\n#a</text>")

	doc, err := Parse(r)

	if len(doc) != 1 || err != nil {
		t.Errorf("Document: %s. Error: %s", doc, err)
		t.FailNow()
	}

	pair := Pair{"someKey", "#c\n#someValue\nd\n#a"}

	if len(doc) == 1 && doc[0] != pair {
		t.Errorf("Expected %s. Got %s", pair, doc[0])
	}
}
