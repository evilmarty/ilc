package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

func assertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s\nExpected: %v\nActual: %v", msg, expected, actual)
	}
}

func assertDeepEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		b := strings.Builder{}
		pretty.Fdiff(&b, expected, actual)
		t.Errorf("%s\nExpected: %v\nActual: %v\nDiff: %s", msg, expected, actual, b.String())
	}
}

func fatalDiff(t *testing.T, expected, actual interface{}) {
	t.Helper()
	b := strings.Builder{}
	pretty.Fdiff(&b, expected, actual)
	t.Fatal(b.String())
}
