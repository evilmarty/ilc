package main

import "testing"

func TestScriptCommand(t *testing.T) {
	cmd := ScriptCommand("echo foobar")
	cmd.Stdout = nil
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if string(output) != "foobar\n" {
		t.Errorf("Unexpected output: %s", output)
	}
}
