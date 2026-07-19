package tui

import (
	"reflect"
	"testing"
)

// feed applies keys and returns the final state flags.
func feed(s *selector, keys string) (done, canceled bool) {
	for i := 0; i < len(keys); i++ {
		done, canceled = s.handle(keys[i])
		if done {
			return done, canceled
		}
	}
	return done, canceled
}

func TestSelectorDefaultsToAll(t *testing.T) {
	s := newSelector(3)
	done, canceled := feed(s, "\r")
	if !done || canceled {
		t.Fatalf("enter: done=%v canceled=%v", done, canceled)
	}
	if got := s.picked(); !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Errorf("picked = %v, want all", got)
	}
}

func TestSelectorToggleAndMove(t *testing.T) {
	s := newSelector(3)
	feed(s, "j ") // down to 1, deselect it
	if got := s.picked(); !reflect.DeepEqual(got, []int{0, 2}) {
		t.Errorf("picked = %v, want [0 2]", got)
	}

	// cursor is clamped at both ends
	feed(s, "kkkk")
	if s.cursor != 0 {
		t.Errorf("cursor = %d, want 0", s.cursor)
	}
	feed(s, "jjjjjj")
	if s.cursor != 2 {
		t.Errorf("cursor = %d, want 2", s.cursor)
	}
}

func TestSelectorToggleAll(t *testing.T) {
	s := newSelector(2)
	s.handle('a') // all selected -> none
	if got := s.picked(); got != nil {
		t.Errorf("picked = %v, want none", got)
	}
	s.handle('a') // none -> all
	if got := s.picked(); !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("picked = %v, want all", got)
	}
}

func TestSelectorCancel(t *testing.T) {
	for _, key := range []byte{'q', 3} { // q, Ctrl-C
		s := newSelector(2)
		done, canceled := s.handle(key)
		if !done || !canceled {
			t.Errorf("key %d: done=%v canceled=%v, want cancel", key, done, canceled)
		}
	}
}
