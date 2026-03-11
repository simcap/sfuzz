package assert

import (
	"bytes"
	"testing"
)

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("\n got: %v\nwant: %v\n", got, want)
	}
}

func EqualBytes(t *testing.T, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("\n got: %q\nwant: %q\n", got, want)
	}
}
