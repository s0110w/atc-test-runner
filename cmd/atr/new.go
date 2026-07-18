package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atr/internal/atcoder"
	"atr/internal/config"
)

// taskDirName is the task ID's suffix ("abc300_a" -> "a") for fast cd;
// the full ID is recoverable as parentDir + "_" + name.
func taskDirName(task string, used map[string]bool) string {
	short := task
	if i := strings.LastIndex(task, "_"); i >= 0 && i+1 < len(task) {
		short = task[i+1:]
	}
	if used[short] {
		short = task // suffix collision in an irregular contest: keep the full ID
	}
	used[short] = true
	return short
}

func cmdNew(args []string) error {
	if len(args) != 1 || !atcoder.IsContestID(args[0]) {
		return fmt.Errorf("%w: atr new <contest ID (e.g. abc300)>", errUsage)
	}
	contest := args[0]
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	body, err := atcoder.FetchPage(atcoder.ContestTasksURL(contest))
	if err != nil {
		return err
	}
	tasks := atcoder.ExtractTasks(body, contest)
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found for %s (during a contest, set ATR_SESSION to your REVEL_SESSION cookie)", contest)
	}

	used := map[string]bool{}
	for _, task := range tasks {
		dir := filepath.Join(contest, taskDirName(task, used))
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("skip: %s (already exists)\n", dir)
			continue
		}
		if err := downloadTask(atcoder.TaskPageURL(contest, task), filepath.Join(dir, "test")); err != nil {
			return err
		}
		if cfg.Template != "" {
			if err := os.CopyFS(dir, os.DirFS(cfg.Template)); err != nil {
				return fmt.Errorf("copy template: %v", err)
			}
		}
	}
	return nil
}
