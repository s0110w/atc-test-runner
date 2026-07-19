// Package atcoder accesses atcoder.jp and parses its pages.
package atcoder

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// UserAgent is set by the CLI layer to "atr/VERSION".
var UserAgent = "atr"

var (
	// ponytail: regex over AtCoder's stable markup; switch to golang.org/x/net/html if a page breaks it
	sampleRe    = regexp.MustCompile(`(?s)<h3[^>]*>\s*(入力例|出力例|Sample Input|Sample Output)\s*(\d+)\s*</h3>\s*(?:<div[^>]*>\s*)?<pre[^>]*>(.*?)</pre>`)
	problemIDRe = regexp.MustCompile(`^([a-z0-9]+)_[a-z0-9]+$`)
	contestIDRe = regexp.MustCompile(`^[a-z0-9_-]+$`)
	taskLinkRe  = regexp.MustCompile(`href="/contests/([a-z0-9_-]+)/tasks/([a-z0-9_-]+)"`)
)

func IsContestID(s string) bool { return contestIDRe.MatchString(s) }

func ContestTasksURL(contest string) string {
	return "https://atcoder.jp/contests/" + contest + "/tasks"
}

func TaskPageURL(contest, task string) string {
	return fmt.Sprintf("https://atcoder.jp/contests/%s/tasks/%s", contest, task)
}

// TaskURL accepts a full task URL or a problem ID like "abc300_a"
// (the part before "_" is the contest). Irregular IDs need a full URL.
func TaskURL(arg string) (string, error) {
	if strings.Contains(arg, "atcoder.jp/contests/") {
		return arg, nil
	}
	if m := problemIDRe.FindStringSubmatch(arg); m != nil {
		return TaskPageURL(m[1], arg), nil
	}
	return "", fmt.Errorf("expected a task URL or problem ID like abc300_a, got %q", arg)
}

// fetchInterval throttles requests: atr new hits the contest page plus one
// page per task in a tight loop, which AtCoder answers with 429.
// ponytail: fixed interval, no adaptive backoff on 429
const fetchInterval = 1500 * time.Millisecond

var lastFetch time.Time

// FetchPage GETs a page with a 30s timeout and no retry. The ATR_SESSION
// environment variable, if set, is sent as the REVEL_SESSION cookie so
// that pages of a running contest are reachable.
func FetchPage(url string) (string, error) {
	if wait := fetchInterval - time.Since(lastFetch); wait > 0 {
		time.Sleep(wait)
	}
	lastFetch = time.Now()

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	if s := os.Getenv("ATR_SESSION"); s != "" {
		req.AddCookie(&http.Cookie{Name: "REVEL_SESSION", Value: s})
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ExtractSamples returns filename -> content, e.g. "sample-1.in" -> "2\n1 3 1 2\n".
// Content is normalized: leading newline stripped, exactly one trailing newline.
func ExtractSamples(body string) map[string]string {
	samples := map[string]string{}
	for _, m := range sampleRe.FindAllStringSubmatch(body, -1) {
		ext := "in"
		if m[1] == "出力例" || m[1] == "Sample Output" {
			ext = "out"
		}
		name := fmt.Sprintf("sample-%s.%s", m[2], ext)
		if _, ok := samples[name]; ok {
			continue // the English section duplicates the Japanese one
		}
		s := html.UnescapeString(m[3])
		s = strings.TrimPrefix(s, "\n")
		s = strings.TrimRight(s, "\n") + "\n"
		samples[name] = s
	}
	return samples
}

// ExtractTasks returns the contest's task IDs in page order, deduped
// (each task is linked more than once on the tasks page).
func ExtractTasks(body, contest string) []string {
	seen := map[string]bool{}
	var tasks []string
	for _, m := range taskLinkRe.FindAllStringSubmatch(body, -1) {
		if m[1] != contest || seen[m[2]] {
			continue
		}
		seen[m[2]] = true
		tasks = append(tasks, m[2])
	}
	return tasks
}
