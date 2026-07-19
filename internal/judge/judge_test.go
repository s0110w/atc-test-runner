package judge

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeCase creates test/NAME.in (and .out unless out is "-") and
// returns the .in path.
func writeCase(t *testing.T, dir, in, out string) string {
	t.Helper()
	inPath := filepath.Join(dir, "case.in")
	if err := os.WriteFile(inPath, []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	if out != "-" {
		if err := os.WriteFile(filepath.Join(dir, "case.out"), []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return inPath
}

// RunCase is exercised with real processes: this is the heart of the
// tool, and its failure modes (hangs, orphaned children) only show up
// with actual process execution.
func TestRunCaseVerdicts(t *testing.T) {
	cases := []struct {
		name    string
		command string
		in, out string
		tle     float64
		want    bool
	}{
		{"AC", "cat", "5\n", "5\n", 0, true},
		{"AC missing trailing newline", `printf 5`, "", "5\n", 0, true},
		{"WA", "echo 4", "", "5\n", 0, false},
		{"RE beats matching output", "echo 5; exit 3", "", "5\n", 0, false},
		{"RE command not found", "definitely-not-a-command-xyz", "", "5\n", 0, false},
		{"no expected output", "cat", "5\n", "-", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := writeCase(t, t.TempDir(), c.in, c.out)
			if got := RunCase(c.command, in, c.tle); got != c.want {
				t.Errorf("RunCase(%q) = %v, want %v", c.command, got, c.want)
			}
		})
	}
}

// A pipeline that ignores the timeout must still return promptly:
// the whole process group is killed, not just sh.
func TestRunCaseTLEKillsPipeline(t *testing.T) {
	in := writeCase(t, t.TempDir(), "", "x\n")
	start := time.Now()
	if RunCase("sleep 5 | cat", in, 0.3) {
		t.Error("should not be AC")
	}
	if e := time.Since(start); e > 3*time.Second {
		t.Errorf("TLE took %v; pipeline children not killed", e)
	}
}

// Unbounded output must be cut off and judged RE quickly, even with no
// time limit: the child is killed when the cap is hit, otherwise it
// would block on a full pipe forever.
func TestRunCaseOutputLimit(t *testing.T) {
	old := maxOutputBytes
	maxOutputBytes = 1 << 20
	defer func() { maxOutputBytes = old }()

	in := writeCase(t, t.TempDir(), "", "x\n")
	start := time.Now()
	if RunCase("cat /dev/zero", in, 0) {
		t.Error("should not be AC")
	}
	if e := time.Since(start); e > 5*time.Second {
		t.Errorf("output-limit RE took %v; child not killed on overflow", e)
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
