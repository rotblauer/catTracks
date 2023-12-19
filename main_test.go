package main

import (
	"testing"
)

func TestBashExecLog(t *testing.T) {
	err := bashExec("echo 'hello world'", "abc ")
	if err != nil {
		t.Fatal(err)
	}

}
