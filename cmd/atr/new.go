package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atr/internal/atcoder"
	"atr/internal/config"
	"atr/internal/tui"
)

// contest.json gives tools and humans the per-task metadata needed to
// work a contest: titles, task page URLs and submit URLs.
type taskMeta struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Title     string `json:"title"`
	Dir       string `json:"dir"`
	URL       string `json:"url"`
	SubmitURL string `json:"submit_url"`
}

type contestMeta struct {
	Contest string     `json:"contest"`
	URL     string     `json:"url"`
	Tasks   []taskMeta `json:"tasks"`
}

func writeContestJSON(contest string, tasks []atcoder.Task, dirNames []string) error {
	meta := contestMeta{Contest: contest, URL: atcoder.ContestURL(contest)}
	for i, t := range tasks {
		meta.Tasks = append(meta.Tasks, taskMeta{
			ID:        t.ID,
			Label:     t.Label,
			Title:     t.Title,
			Dir:       dirNames[i],
			URL:       atcoder.TaskPageURL(contest, t.ID),
			SubmitURL: atcoder.SubmitURL(contest, t.ID),
		})
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(contest, "contest.json"), append(data, '\n'), 0o644)
}

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
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	sel := fs.Bool("s", false, "select tasks interactively before setup")
	fs.Parse(args)
	if fs.NArg() != 1 || !atcoder.IsContestID(fs.Arg(0)) {
		return fmt.Errorf("%w: atr new [-s] <contest ID (e.g. abc300)>", errUsage)
	}
	contest := fs.Arg(0)
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

	// dir names are computed over all tasks first so that collision
	// handling does not depend on what gets selected
	used := map[string]bool{}
	dirNames := make([]string, len(tasks))
	for i, task := range tasks {
		dirNames[i] = taskDirName(task.ID, used)
	}

	picked := make([]int, len(tasks))
	for i := range picked {
		picked[i] = i
	}
	if *sel || cfg.Select {
		picked, err = tui.SelectTasks(contest, dirNames)
		if err != nil {
			return err
		}
		if len(picked) == 0 {
			fmt.Println("nothing selected")
			return nil
		}
	}

	// the contest template is expanded once, only when the contest dir
	// is created by this run
	if _, err := os.Stat(contest); os.IsNotExist(err) {
		if err := os.MkdirAll(contest, 0o755); err != nil {
			return err
		}
		if cfg.ContestTemplate != "" {
			if err := os.CopyFS(contest, os.DirFS(cfg.ContestTemplate)); err != nil {
				return fmt.Errorf("copy contest template: %v", err)
			}
		}
	} else if cfg.ContestTemplate != "" {
		fmt.Printf("skip contest template: %s/ already exists\n", contest)
	}

	// contest.json always covers every task and is regenerated on each
	// run: it is derived metadata, not user data, so overwriting is safe
	if err := writeContestJSON(contest, tasks, dirNames); err != nil {
		return err
	}
	fmt.Printf("saved: %s\n", filepath.Join(contest, "contest.json"))

	for _, i := range picked {
		dir := filepath.Join(contest, dirNames[i])
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("skip: %s (already exists)\n", dir)
			continue
		}
		if err := downloadTask(atcoder.TaskPageURL(contest, tasks[i].ID), filepath.Join(dir, "test")); err != nil {
			return err
		}
		if cfg.TaskTemplate != "" {
			if err := os.CopyFS(dir, os.DirFS(cfg.TaskTemplate)); err != nil {
				return fmt.Errorf("copy task template: %v", err)
			}
		}
	}
	return nil
}
