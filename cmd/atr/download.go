package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"atr/internal/atcoder"
)

// downloadTask fetches a task page and saves its samples into testDir.
func downloadTask(url, testDir string) error {
	body, err := atcoder.FetchPage(url)
	if err != nil {
		return err
	}
	samples := atcoder.ExtractSamples(body)
	if len(samples) == 0 {
		return fmt.Errorf("no samples found at %s (during a contest, set ATR_SESSION to your REVEL_SESSION cookie)", url)
	}
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return err
	}
	names := make([]string, 0, len(samples))
	for name := range samples {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(testDir, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("%v (remove %s/ to re-download)", err, testDir)
		}
		if _, err := f.WriteString(samples[name]); err != nil {
			f.Close()
			return err
		}
		f.Close()
		fmt.Printf("saved: %s\n", path)
	}
	return nil
}

func cmdDownload(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: atr download <URL or problem ID>", errUsage)
	}
	url, err := atcoder.TaskURL(args[0])
	if err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	return downloadTask(url, "test")
}
