package test

import (
	"testing"
)

func TestTrivial(t *testing.T) {
	if !true {
		t.Error("Something went wrong in trivial test")
	}
}
