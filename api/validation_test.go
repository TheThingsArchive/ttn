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

func (i *invalid) Validate() error {
	return NotNilAndValid(i.nestedPtr, "nestedPtr")
}

func TestNotNilAndValid(t *testing.T) {
	subject := invalid{}
	err := NotNilAndValid(subject.nestedPtr, "subject")
	if err == nil || err.Error() != "subject not valid: can not be empty" {
		t.Error("Expected validation error: 'subject not valid: can not be empty' but found", err)
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
