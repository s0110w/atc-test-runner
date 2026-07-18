package atcoder

import (
	"reflect"
	"testing"
)

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
	got := ExtractSamples(taskHTML)
	want := map[string]string{
		"sample-1.in":  "2\n1 3 1 2\n",
		"sample-1.out": "3\n",
		"sample-2.in":  "1 < 2 & 3\n", // entities unescaped, ja/en deduped
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractSamples = %v, want %v", got, want)
	}
}

const tasksHTML = `
<table><tbody>
<tr><td><a href="/contests/abc300/tasks/abc300_a">A</a></td>
    <td><a href="/contests/abc300/tasks/abc300_a">Story</a></td></tr>
<tr><td><a href="/contests/abc300/tasks/abc300_b">B</a></td></tr>
<tr><td><a href="/contests/other/tasks/other_a">unrelated</a></td></tr>
</tbody></table>
`

func TestExtractTasks(t *testing.T) {
	got := ExtractTasks(tasksHTML, "abc300")
	want := []string{"abc300_a", "abc300_b"} // deduped, page order, other contests excluded
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractTasks = %v, want %v", got, want)
	}
}

func TestTaskURL(t *testing.T) {
	for arg, want := range map[string]string{
		"abc300_a": "https://atcoder.jp/contests/abc300/tasks/abc300_a",
		"https://atcoder.jp/contests/abc086/tasks/abc086_a": "https://atcoder.jp/contests/abc086/tasks/abc086_a",
	} {
		got, err := TaskURL(arg)
		if err != nil || got != want {
			t.Errorf("TaskURL(%q) = %q, %v; want %q", arg, got, err, want)
		}
	}
	if _, err := TaskURL("not a task"); err == nil {
		t.Error("TaskURL should reject invalid input")
	}
}
