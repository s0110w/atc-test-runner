package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"atr/internal/atcoder"
)

func TestWriteContestJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("abc300", 0o755); err != nil {
		t.Fatal(err)
	}
	tasks := []atcoder.Task{{ID: "abc300_a", Label: "A", Title: "Story"}}
	if err := writeContestJSON("abc300", tasks, []string{"a"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join("abc300", "contest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got contestMeta
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := contestMeta{
		Contest: "abc300",
		URL:     "https://atcoder.jp/contests/abc300",
		Tasks: []taskMeta{{
			ID: "abc300_a", Label: "A", Title: "Story", Dir: "a",
			URL:       "https://atcoder.jp/contests/abc300/tasks/abc300_a",
			SubmitURL: "https://atcoder.jp/contests/abc300/submit?taskScreenName=abc300_a",
		}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("contest.json = %+v, want %+v", got, want)
	}
}
