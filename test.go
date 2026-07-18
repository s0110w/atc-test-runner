package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	maxOutputBytes  = 64 << 20 // captured stdout beyond this is RE: buggy solutions can print unbounded output
	maxDisplayLines = 40
)

func green(s string) string { return "\033[32m" + s + "\033[0m" }
func red(s string) string   { return "\033[31m" + s + "\033[0m" }

func cmdTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	command := fs.String("c", "./a.out", "command to test")
	tle := fs.Float64("t", 10, "time limit in seconds (0: no limit)")
	dir := fs.String("d", "test", "test case directory")
	fs.Parse(args)

	ins, err := filepath.Glob(filepath.Join(*dir, "*.in"))
	if err != nil || len(ins) == 0 {
		return fmt.Errorf("no test cases found in %s/", *dir)
	}
	sort.Strings(ins)
	fmt.Printf("%d cases found\n", len(ins))

	ac := 0
	for _, in := range ins {
		if runCase(*command, in, *tle) {
			ac++
		}
	}
	fmt.Println()
	if ac == len(ins) {
		fmt.Printf("test %s: %d cases\n", green("success"), len(ins))
		return nil
	}
	return fmt.Errorf("test %s: %d AC / %d cases", red("failed"), ac, len(ins))
}

// capWriter aborts the capture once the output exceeds maxOutputBytes,
// which surfaces as an error from cmd.Wait (judged RE).
type capWriter struct {
	buf bytes.Buffer
}

func (w *capWriter) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > maxOutputBytes {
		return 0, fmt.Errorf("output exceeds %d bytes", maxOutputBytes)
	}
	return w.buf.Write(p)
}

func runCase(command, inPath string, tle float64) bool {
	fmt.Printf("\n%s\n", strings.TrimSuffix(filepath.Base(inPath), ".in"))

	expected, err := os.ReadFile(strings.TrimSuffix(inPath, ".in") + ".out")
	if err != nil {
		// an unjudged case must not count as success: exit code 0 means "all cases AC"
		fmt.Println(red("no expected output file; cannot judge"))
		return false
	}
	inFile, err := os.Open(inPath)
	if err != nil {
		fmt.Println(red("error:"), err)
		return false
	}
	defer inFile.Close()

	ctx := context.Background()
	if tle > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(tle*float64(time.Second)))
		defer cancel()
	}

	// ponytail: sh -c, no Windows support (a design non-goal)
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdin = inFile
	var stdout capWriter
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	// kill the whole process group on timeout: killing only sh leaves
	// pipeline children holding stdout, and Wait would block forever
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second

	start := time.Now()
	err = cmd.Run()
	fmt.Printf("time: %.6f sec\n", time.Since(start).Seconds())

	// TLE means "the process was killed for exceeding the limit", not
	// "the deadline has passed by now": a run finishing just under the
	// limit must not turn into TLE while we judge it
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		fmt.Println(red("TLE"))
		return false
	}
	if err != nil {
		// nonzero exit is RE even if the output matches: a crash that
		// happens to print the right answer must not pass (soundness)
		fmt.Printf("%s: %v\n", red("RE"), err)
		return false
	}

	if outputsMatch(stdout.buf.Bytes(), expected) {
		fmt.Println(green("AC"))
		return true
	}
	fmt.Println(red("WA"))
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
