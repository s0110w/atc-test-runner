package judge

import "testing"

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
