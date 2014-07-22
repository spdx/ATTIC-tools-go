package spdx

import "testing"

import (
	"strings"
)

func TestValueCreator(t *testing.T) {
	c := NewValueCreator("what: who (email@example.org)", nil)
	if c.What() != "what" {
		t.Errorf("Wrong what %#v", c.What())
	}
	if c.Name() != "who" {
		t.Errorf("Wrong who %#v", c.Name())
	}
	if c.Email() != "email@example.org" {
		t.Errorf("Wrong email %#v", c.Email())
	}

	t.Log("No email tests.")
	c = NewValueCreator("what: who", nil)
	if c.What() != "what" {
		t.Errorf("Wrong what %#v", c.What())
	}
	if c.Name() != "who" {
		t.Errorf("Wrong who %#v", c.Name())
	}
	if c.Email() != "" {
		t.Errorf("Wrong email %#v", c.Email())
	}

	c = NewValueCreator("Incorrect syntax.", nil)
	if c.What() != "" || c.Name() != "" || c.Email() != "" {
		t.Error("Incorrect syntax was parsed.")
	}
	if c.V() != "Incorrect syntax." {
		t.Errorf("Incorrect value for incorrect syntax %#v", c.V())
	}
}

func TestJoinValueStr(t *testing.T) {
	strs := []string{"a", "b", "c", "d"}
	valStrs := make([]ValueStr, len(strs))
	for i := range valStrs {
		valStrs[i] = Str(strs[i], nil)
	}
	strings_join := strings.Join(strs, ",")
	valstr_join := Join(valStrs, ",")
	if strings_join != valstr_join {
		t.Log(valstr_join)
		t.Fail()
	}
}
