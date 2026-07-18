package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ponytail: regex over AtCoder's stable markup; switch to golang.org/x/net/html if a page breaks it
var sampleRe = regexp.MustCompile(`(?s)<h3[^>]*>\s*(入力例|出力例|Sample Input|Sample Output)\s*(\d+)\s*</h3>\s*(?:<div[^>]*>\s*)?<pre[^>]*>(.*?)</pre>`)

var problemIDRe = regexp.MustCompile(`^([a-z0-9]+)_[a-z0-9]+$`)

// taskURL accepts a full task URL or a problem ID like "abc300_a"
// (the part before "_" is the contest). Irregular IDs need a full URL.
func taskURL(arg string) (string, error) {
	if strings.Contains(arg, "atcoder.jp/contests/") {
		return arg, nil
	}
	if m := problemIDRe.FindStringSubmatch(arg); m != nil {
		return fmt.Sprintf("https://atcoder.jp/contests/%s/tasks/%s", m[1], arg), nil
	}
	return "", fmt.Errorf("%w: expected a task URL or problem ID like abc300_a, got %q", errUsage, arg)
}

// extractSamples returns filename -> content, e.g. "sample-1.in" -> "2\n1 3 1 2\n".
// Content is normalized: leading newline stripped, exactly one trailing newline.
func extractSamples(body string) map[string]string {
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

func cmdDownload(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: atr download <URL or problem ID>", errUsage)
	}
	url, err := taskURL(args[0])
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "atr/"+version)
	if s := os.Getenv("ATR_SESSION"); s != "" {
		req.AddCookie(&http.Cookie{Name: "REVEL_SESSION", Value: s})
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	samples := extractSamples(string(body))
	if len(samples) == 0 {
		return fmt.Errorf("no samples found at %s (during a contest, set ATR_SESSION to your REVEL_SESSION cookie)", url)
	}
	if err := os.MkdirAll("test", 0o755); err != nil {
		return err
	}
	names := make([]string, 0, len(samples))
	for name := range samples {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join("test", name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("%v (remove test/ to re-download)", err)
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
