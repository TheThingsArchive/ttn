package api

import (
	"testing"
)

func TestValidate_nil(t *testing.T) {
	Validate(nil)
}

type invalid struct {
	nestedPtr *interface{}
}

func (i invalid) Validate() error {
	return NotNilAndValid(i.nestedPtr, "nestedPtr")
}

func TestNotNilAndValidStruct(t *testing.T) {
	subject := invalid{}

	err := NotNilAndValid(subject, "subject")
	if err == nil || err.Error() != "Invalid subject: nestedPtr not valid: can not be empty" {
		t.Error("Expected validation error: 'Invalid subject: nestedPtr not valid: can not be empty", err)
	}
}

func TestNotNilAndValidPtr(t *testing.T) {
	subject := &invalid{}
	err := NotNilAndValid(subject, "subject")
	if err == nil || err.Error() != "Invalid subject: nestedPtr not valid: can not be empty" {
		t.Error("Expected validation error: 'Invalid subject: nestedPtr not valid: can not be empty", err)
	}
}

type testInterface interface {
	Nothing()
}

func TestNotNilAndValidDifferentTypes(t *testing.T) {
	err := NotNilAndValid(struct{}{}, "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}

	err = NotNilAndValid(&struct{}{}, "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}

	err = NotNilAndValid(func() {}, "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}

	err = NotNilAndValid(make(chan byte), "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}

	err = NotNilAndValid(make([]byte, 0), "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}

	err = NotNilAndValid(make(map[byte]byte, 0), "subject")
	if err != nil {
		t.Error("Expected nil but got", err)
	}
}

func TestValidID(t *testing.T) {
	if ValidID("a") {
		t.Error("'a' is not a valid id")
	}
}

func TestNotEmptyAndValidId(t *testing.T) {
	err := NotEmptyAndValidId("", "subject")
	if err == nil || err.Error() != "subject not valid: can not be empty" {
		t.Error("Expected validation error: 'subject not valid: can not be empty' but found", err)
	}

	err = NotEmptyAndValidId("a", "subject")
	if err == nil || err.Error() != "subject not valid: has wrong format a" {
		t.Error("Expected validation error: 'subject not valid: has wrong format a' but found", err)
	}
}
