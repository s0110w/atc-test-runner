package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"atr/internal/config"
	"atr/internal/judge"
)

func cmdTest(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	command := fs.String("c", cfg.Command, "command to test")
	build := fs.String("b", cfg.Build, "build command, run once before the cases")
	tle := fs.Float64("t", 10, "time limit in seconds (0: no limit)")
	dir := fs.String("d", "test", "test case directory")
	fs.Parse(args)

	run := expandTask(*command)
	if run == "./a.out" && *build == "" {
		if _, err := os.Stat("./a.out"); err != nil {
			return fmt.Errorf(`run command is not set and ./a.out does not exist — build it first, pass -c, or set "command" in atr.toml`)
		}
	}

	ins, err := filepath.Glob(filepath.Join(*dir, "*.in"))
	if err != nil || len(ins) == 0 {
		return fmt.Errorf("no test cases found in %s/", *dir)
	}
	sort.Strings(ins)

	if b := expandTask(*build); b != "" {
		fmt.Printf("$ %s\n", b)
		bc := exec.Command("sh", "-c", b)
		bc.Stdout, bc.Stderr = os.Stdout, os.Stderr
		if err := bc.Run(); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}
	}
	fmt.Printf("%d cases found\n", len(ins))

	ac := 0
	for _, in := range ins {
		if judge.RunCase(run, in, *tle) {
			ac++
		}
	}
	fmt.Println()
	if ac == len(ins) {
		fmt.Printf("test %s: %d cases\n", judge.Green("success"), len(ins))
		return nil
	}
	return fmt.Errorf("test %s: %d AC / %d cases", judge.Red("failed"), ac, len(ins))
}

// expandTask replaces {task} with the working directory's name, so one
// shared config can address per-task build targets (e.g. "cabal run {task}").
func expandTask(s string) string {
	wd, err := os.Getwd()
	if err != nil {
		return s
	}
	return strings.ReplaceAll(s, "{task}", filepath.Base(wd))
}
