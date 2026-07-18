// Package config finds and parses the nearest atr.toml.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Command  string // default command for `atr test -c`
	Template string // absolute path of the template dir, "" if unset
}

func Default() Config { return Config{Command: "./a.out"} }

// Load uses the nearest atr.toml from the working directory upward
// (including the working directory itself). Without one, every command
// still works on hardcoded defaults.
func Load() (Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return Default(), nil
	}
	return Find(dir)
}

func Find(dir string) (Config, error) {
	for {
		data, err := os.ReadFile(filepath.Join(dir, "atr.toml"))
		if err == nil {
			return parse(string(data), dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Default(), nil
		}
		dir = parent
	}
}

// parse reads a minimal TOML subset: `key = "string"` lines and
// # comments. Anything outside the subset is an error, never a silent
// misread of valid TOML.
// ponytail: flat string keys only; adopt a TOML library if structure is ever needed
func parse(text, dir string) (Config, error) {
	cfg := Default()
	for i, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rest, ok := strings.Cut(line, "=")
		if !ok {
			return cfg, fmt.Errorf(`atr.toml:%d: only 'key = "string"' lines and # comments are supported`, i+1)
		}
		val, err := parseTOMLString(strings.TrimSpace(rest))
		if err != nil {
			return cfg, fmt.Errorf("atr.toml:%d: %v", i+1, err)
		}
		switch strings.TrimSpace(key) {
		case "command":
			cfg.Command = val
		case "template":
			if !filepath.IsAbs(val) {
				val = filepath.Join(dir, val) // relative to the config file's directory
			}
			cfg.Template = val
		default:
			return cfg, fmt.Errorf("atr.toml:%d: unknown key %q", i+1, strings.TrimSpace(key))
		}
	}
	return cfg, nil
}

// parseTOMLString accepts a basic double-quoted string without escapes,
// optionally followed by a # comment.
func parseTOMLString(s string) (string, error) {
	if len(s) < 2 || s[0] != '"' {
		return "", fmt.Errorf("value must be a double-quoted string")
	}
	end := strings.IndexByte(s[1:], '"')
	if end < 0 {
		return "", fmt.Errorf("unclosed string")
	}
	val := s[1 : end+1]
	if strings.ContainsRune(val, '\\') {
		return "", fmt.Errorf("escape sequences are not supported")
	}
	rest := strings.TrimSpace(s[end+2:])
	if rest != "" && !strings.HasPrefix(rest, "#") {
		return "", fmt.Errorf("unexpected content after string: %q", rest)
	}
	return val, nil
}
