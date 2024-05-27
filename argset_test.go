package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestParseArgSet(t *testing.T) {
	expected := ArgSet{
		Commands: []string{},
		Params:   map[string]string{},
	}
	args := []string{}
	if actual := ParseArgSet(args); !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v to match %v", expected, actual)
	}

	expected = ArgSet{
		Commands: []string{"foo", "bar", "baz"},
		Params: map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}
	args = []string{"foo", "-a", "1", "bar", "--b", "2", "baz", "--c=3"}
	if actual := ParseArgSet(args); !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v to match %v", expected, actual)
	}
}

func TestArgsetParamNames(t *testing.T) {
	argset := ArgSet{
		Params: map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}
	expected := []string{"a", "b", "c"}
	actual := argset.ParamNames()
	sort.Strings(actual)
	assertDeepEqual(t, expected, actual, "ArgSet.ParamNames() returned unexpected result")
}
