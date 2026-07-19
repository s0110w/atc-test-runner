package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandTask(t *testing.T) {
	t.Chdir(t.TempDir())
	wd, _ := os.Getwd()
	task := filepath.Base(wd)

	if got, want := expandTask("cabal run {task} -v0"), "cabal run "+task+" -v0"; got != want {
		t.Errorf("expandTask = %q, want %q", got, want)
	}
	if got := expandTask("./a.out"); got != "./a.out" {
		t.Errorf("no placeholder: got %q", got)
	}
}
