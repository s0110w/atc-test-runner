package main

import "testing"

func TestTaskDirName(t *testing.T) {
	used := map[string]bool{}
	if got := taskDirName("abc300_a", used); got != "a" {
		t.Errorf("got %q, want a", got)
	}
	if got := taskDirName("weird_a", used); got != "weird_a" {
		t.Errorf("suffix collision should keep the full ID, got %q", got)
	}
	if got := taskDirName("nounderscore", used); got != "nounderscore" {
		t.Errorf("got %q, want nounderscore", got)
	}
}
