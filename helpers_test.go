package tracer

import (
	"reflect"
	"testing"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func assertTrue(t *testing.T, value bool) {
	t.Helper()
	if !value {
		t.Fatalf("expected true, got false")
	}
}

func assertFalse(t *testing.T, value bool) {
	t.Helper()
	if value {
		t.Fatalf("expected false, got true")
	}
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
