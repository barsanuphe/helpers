package helpers

import (
	"fmt"
	"testing"
)

func TestHelpersStringInSlice(t *testing.T) {
	fmt.Println("+ Testing Helpers/StringInSlice()...")
	candidates := []string{"one", "two"}
	idx, isIn := StringInSlice("one", candidates)
	if !isIn || idx != 0 {
		t.Error("Error finding string in slice")
	}
	idx, isIn = StringInSlice("One", candidates)
	if isIn || idx != -1 {
		t.Error("Error finding string in slice")
	}
}

func TestHelpersCSContains(t *testing.T) {
	fmt.Println("+ Testing Helpers/CaseInsensitiveContains()...")
	if !CaseInsensitiveContains("TestString", "test") {
		t.Error("Error, substring in string")
	}
	if !CaseInsensitiveContains("TestString", "stSt") {
		t.Error("Error, substring in string")
	}
	if CaseInsensitiveContains("TestString", "teest") {
		t.Error("Error, substring not in string")
	}
}
