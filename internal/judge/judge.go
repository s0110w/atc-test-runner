// Package judge runs one test case against a command and judges the result.
package judge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// maxOutputBytes is a variable so tests can lower it.
var maxOutputBytes = 64 << 20 // captured stdout beyond this is RE: buggy solutions can print unbounded output

const maxDisplayLines = 40

func Green(s string) string { return "\033[32m" + s + "\033[0m" }
func Red(s string) string   { return "\033[31m" + s + "\033[0m" }

// capWriter aborts the capture once the output exceeds the limit and
// fires onOver. Killing the process there is essential: merely stopping
// the capture leaves the child blocked on a full pipe, and Wait would
// then hang until the deadline (or forever with no time limit).
type capWriter struct {
	buf    bytes.Buffer
	limit  int
	onOver func()
	over   bool
}

func (w *capWriter) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > w.limit {
		if !w.over {
			w.over = true
			if w.onOver != nil {
				w.onOver()
			}
		}
		return 0, fmt.Errorf("output exceeds %d bytes", w.limit)
	}
	return w.buf.Write(p)
}

// RunCase runs command with the .in file on stdin and prints the verdict.
// It returns true only for AC.
func RunCase(command, inPath string, tleSeconds float64) bool {
	fmt.Printf("\n%s\n", strings.TrimSuffix(filepath.Base(inPath), ".in"))

	expected, err := os.ReadFile(strings.TrimSuffix(inPath, ".in") + ".out")
	if err != nil {
		// an unjudged case must not count as success: exit code 0 means "all cases AC"
		fmt.Println(Red("no expected output file; cannot judge"))
		return false
	}
	inFile, err := os.Open(inPath)
	if err != nil {
		fmt.Println(Red("error:"), err)
		return false
	}
	defer inFile.Close()

	ctx := context.Background()
	if tleSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(tleSeconds*float64(time.Second)))
		defer cancel()
	}

	// ponytail: sh -c, no Windows support (a design non-goal)
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdin = inFile
	stdout := &capWriter{limit: maxOutputBytes}
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	// kill the whole process group on timeout: killing only sh leaves
	// pipeline children holding stdout, and Wait would block forever
	killGroup := func() { syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { killGroup(); return nil }
	cmd.WaitDelay = time.Second
	stdout.onOver = killGroup

	start := time.Now()
	err = cmd.Run()
	fmt.Printf("time: %.6f sec\n", time.Since(start).Seconds())

	if stdout.over {
		fmt.Printf("%s: output limit exceeded (%d bytes)\n", Red("RE"), stdout.limit)
		return false
	}
	// TLE means "the process was killed for exceeding the limit", not
	// "the deadline has passed by now": a run finishing just under the
	// limit must not turn into TLE while we judge it
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		fmt.Println(Red("TLE"))
		return false
	}
	if err != nil {
		// nonzero exit is RE even if the output matches: a crash that
		// happens to print the right answer must not pass (soundness)
		fmt.Printf("%s: %v\n", Red("RE"), err)
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 127 {
			fmt.Println("(exit status 127 usually means the run command was not found — check -c or \"command\" in atr.toml)")
		}
		return false
	}

	if outputsMatch(stdout.buf.Bytes(), expected) {
		fmt.Println(Green("AC"))
		return true
	}
	fmt.Println(Red("WA"))
	input, _ := os.ReadFile(inPath)
	printSection("input", input)
	printSection("output", stdout.buf.Bytes())
	printSection("expected", expected)
	return false
}

// outputsMatch is exact match, except CRLF-insensitive and tolerant of
// trailing newlines (AtCoder's judge accepts a missing final newline,
// so rejecting it here would fail correct printf-style solutions).
func outputsMatch(actual, expected []byte) bool {
	norm := func(b []byte) string {
		return strings.TrimRight(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	}
	return norm(actual) == norm(expected)
}

func printSection(label string, content []byte) {
	fmt.Println(label + ":")
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) > maxDisplayLines {
		lines = append(lines[:maxDisplayLines], fmt.Sprintf("... (%d more lines)", len(lines)-maxDisplayLines))
	}
	fmt.Println(strings.Join(lines, "\n"))
}
