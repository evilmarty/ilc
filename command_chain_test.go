package main

import (
	"reflect"
	"testing"
)

func TestCommandChainLast(t *testing.T) {
	first := &Command{Name: "first"}
	last := &Command{Name: "last"}
	commands := CommandChain{first, last}

	actual := commands.Last()
	if !reflect.DeepEqual(actual, last) {
		t.Fatalf("Expected last element to be %v, not %v", last, actual)
	}
}

func TestCommandChainInputs(t *testing.T) {
	input1 := Input{}
	input2 := Input{}
	commands := CommandChain{
		&Command{
			Name:   "a",
			Inputs: Inputs{input1},
		},
		&Command{
			Name:   "b",
			Inputs: Inputs{input2},
		},
	}

	expected := Inputs{input1, input2}
	actual := commands.Inputs()
	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestCommandChainPure_SingleTrue(t *testing.T) {
	commands := CommandChain{
		&Command{
			Name:   "a",
			Inputs: Inputs{},
			Pure:   true,
		},
	}
	if commands.Pure() != true {
		t.Fatalf("Expected Pure to be %v", true)
	}
}

func TestCommandChainPure_SingleFalse(t *testing.T) {
	commands := CommandChain{
		&Command{
			Name:   "a",
			Inputs: Inputs{},
			Pure:   false,
		},
	}
	if commands.Pure() != false {
		t.Fatalf("Expected Pure to be %v", false)
	}
}

func TestCommandChainPure_NestedTrue(t *testing.T) {
	commands := CommandChain{
		&Command{
			Name:   "a",
			Inputs: Inputs{},
			Pure:   false,
		},
		&Command{
			Name:   "b",
			Inputs: Inputs{},
			Pure:   true,
		},
	}
	if commands.Pure() != true {
		t.Fatalf("Expected Pure to be %v", true)
	}
}

func TestCommandChainPure_NestedFalse(t *testing.T) {
	commands := CommandChain{
		&Command{
			Name:   "a",
			Inputs: Inputs{},
			Pure:   true,
		},
		&Command{
			Name:   "b",
			Inputs: Inputs{},
			Pure:   false,
		},
	}
	if commands.Pure() != false {
		t.Fatalf("Expected Pure to be %v", false)
	}
}
