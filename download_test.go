package main

import "testing"

const taskHTML = `
<span class="lang-ja">
<div class="part"><section><h3>入力例 1</h3><div class="div-btn-copy"><span class="btn-copy">Copy</span></div><pre id="pre-sample0">2
1 3 1 2</pre></section></div>
<div class="part"><section><h3>出力例 1</h3><pre id="pre-sample1">3
</pre></section></div>
</span>
<span class="lang-en">
<div class="part"><section><h3>Sample Input 1</h3><pre>2
1 3 1 2</pre></section></div>
<div class="part"><section><h3>Sample Output 1</h3><pre>3
</pre></section></div>
<div class="part"><section><h3>Sample Input 2</h3><pre>1 &lt; 2 &amp; 3</pre></section></div>
</span>
`

func TestExtractSamples(t *testing.T) {
	got := extractSamples(taskHTML)
	want := map[string]string{
		"sample-1.in":  "2\n1 3 1 2\n",
		"sample-1.out": "3\n",
		"sample-2.in":  "1 < 2 & 3\n", // entities unescaped, ja/en deduped
	}
	if len(got) != len(want) {
		t.Fatalf("got %d samples, want %d: %v", len(got), len(want), got)
	}
	for name, content := range want {
		if got[name] != content {
			t.Errorf("%s = %q, want %q", name, got[name], content)
		}
	}
}

func TestTaskURL(t *testing.T) {
	for arg, want := range map[string]string{
		"abc300_a": "https://atcoder.jp/contests/abc300/tasks/abc300_a",
		"https://atcoder.jp/contests/abc086/tasks/abc086_a": "https://atcoder.jp/contests/abc086/tasks/abc086_a",
	} {
		got, err := taskURL(arg)
		if err != nil || got != want {
			t.Errorf("taskURL(%q) = %q, %v; want %q", arg, got, err, want)
		}
	}
	if _, err := taskURL("not a task"); err == nil {
		t.Error("taskURL should reject invalid input")
	}
}

func TestOutputsMatch(t *testing.T) {
	cases := []struct {
		actual, expected string
		want             bool
	}{
		{"3\n", "3\n", true},
		{"3", "3\n", true},       // missing trailing newline tolerated
		{"3\r\n", "3\n", true},   // CRLF-insensitive
		{"3 \n", "3\n", false},   // trailing space is not tolerated
		{"3\n4\n", "3\n", false}, // extra line
		{" 3\n", "3\n", false},   // leading space
	}
	for _, c := range cases {
		if got := outputsMatch([]byte(c.actual), []byte(c.expected)); got != c.want {
			t.Errorf("outputsMatch(%q, %q) = %v, want %v", c.actual, c.expected, got, c.want)
		}
	}
}
