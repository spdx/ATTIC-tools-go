package rdf

import "testing"

import (
	"errors"
	"github.com/vladvelici/goraptor"
	"github.com/vladvelici/spdx-go/spdx"
)

// Test goraptor term to string
func TestTermStr(t *testing.T) {
	val := "some value"
	if val != termStr(literal(val)) {
		t.Fail()
	}
	if val != termStr(uri(val)) {
		t.Fail()
	}
	if val != termStr(blank(val)) {
		t.Fail()
	}
}

func TestUpd(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.Str("", nil)
	f := upd(&a)
	err := f(blank("world"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.Str("world", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(blank("hello"), nil)
	if err == nil {
		t.Fail()
	}
	if a == spdx.Str("hello", nil) {
		t.Fail()
	}
}

func TestUpdList(t *testing.T) {
	arr := []spdx.ValueStr{spdx.Str("1", nil), spdx.Str("2", nil), spdx.Str("3", nil)}
	f := updList(&arr)
	meta := spdx.NewMeta(5, 7)
	err := f(literal("4"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if len(arr) != 4 || arr[3].Val != "4" || arr[3].M() != meta {
		t.Fail()
	}
}

func TestUpdCutPrefix(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.Str("", nil)
	f := updCutPrefix("prefix_", &a)
	err := f(blank("prefix_world"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.Str("world", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(blank("prefix_hello"), nil)
	if err == nil {
		t.Fail()
	}
	if a == spdx.Str("hello", nil) {
		t.Fail()
	}
}

func TestUpdCreator(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueCreator("", nil)
	f := updCreator(&a)
	err := f(literal("Tool: spdx-go"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.NewValueCreator("Tool: spdx-go", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(literal("Person: Mr. Tester"), nil)
	if err == nil {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestUpdListCreator(t *testing.T) {
	arr := []spdx.ValueCreator{spdx.NewValueCreator("1", nil), spdx.NewValueCreator("2", nil), spdx.NewValueCreator("3", nil)}
	f := updListCreator(&arr)
	meta := spdx.NewMeta(5, 7)
	err := f(literal("4"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if len(arr) != 4 || arr[3].V() != "4" || arr[3].M() != meta {
		t.Fail()
	}
}

func TestUpdDate(t *testing.T) {
	date := "2010-02-03T00:00:00Z"
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueDate("", nil)
	f := updDate(&a)
	err := f(literal(date), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.NewValueDate(date, meta)
	if a.V() == expected.V() && a.Time() == expected.Time() {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(literal("Person: Mr. Tester"), nil)
	if err == nil {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestBuilder(t *testing.T) {
	a := "hello"
	var meta *spdx.Meta
	builder := &builder{t: blank("test"), ptr: &a}
	builder.updaters = map[string]updater{
		"change_value": func(val goraptor.Term, m *spdx.Meta) error {
			a = termStr(val)
			meta = m
			return nil
		},
		"return_error": func(val goraptor.Term, m *spdx.Meta) error {
			return errors.New(termStr(val))
		},
	}
	metaExpected := spdx.NewMeta(3, 4)
	err := builder.apply(uri("change_value"), blank("world"), metaExpected)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if a != "world" {
		t.Errorf("Incorrect update value %#v", metaExpected)
	}
	if meta != metaExpected {
		t.Error("Meta not updated.")
	}
	errorText := "sample error message"
	err = builder.apply(uri("return_error"), blank(errorText), nil)
	if err == nil || err.Error() != errorText {
		t.Errorf("Incorrect error returned %#v", err)
	}
	err = builder.apply(uri("unknown property"), blank(""), nil)
	if err == nil {
		t.Errorf("No error returned for unknown property.")
	}
}
