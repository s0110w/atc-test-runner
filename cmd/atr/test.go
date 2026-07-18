package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"

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
		if judge.RunCase(*command, in, *tle) {
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
