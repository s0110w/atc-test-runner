// Package config finds and parses the nearest atr.toml.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Command         string // default command for `atr test -c`
	Build           string // default build command for `atr test -b`, run once before the cases, "" if unset
	ContestTemplate string // absolute path of the dir expanded into the contest dir, "" if unset
	TaskTemplate    string // absolute path of the dir expanded into each task dir, "" if unset
	Select          bool   // open the task selection TUI by default in `atr new`
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

// parse reads a minimal TOML subset: `key = "string"` / `key = true|false`
// lines and # comments. Anything outside the subset is an error, never a
// silent misread of valid TOML.
// ponytail: flat keys only; adopt a TOML library if structure is ever needed
func parse(text, dir string) (Config, error) {
	cfg := Default()
	abs := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(dir, p) // relative to the config file's directory
	}
	for i, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rest, ok := strings.Cut(line, "=")
		if !ok {
			return cfg, fmt.Errorf(`atr.toml:%d: only 'key = "string"' or 'key = true/false' lines and # comments are supported`, i+1)
		}
		strVal, boolVal, isBool, err := parseValue(strings.TrimSpace(rest))
		if err != nil {
			return cfg, fmt.Errorf("atr.toml:%d: %v", i+1, err)
		}
		key = strings.TrimSpace(key)
		wantBool := key == "select"
		if wantBool != isBool {
			want := "a double-quoted string"
			if wantBool {
				want = "true or false"
			}
			return cfg, fmt.Errorf("atr.toml:%d: %s must be %s", i+1, key, want)
		}
		switch key {
		case "command":
			cfg.Command = strVal
		case "build":
			cfg.Build = strVal
		case "contest_template":
			cfg.ContestTemplate = abs(strVal)
		case "task_template":
			cfg.TaskTemplate = abs(strVal)
		case "select":
			cfg.Select = boolVal
		default:
			return cfg, fmt.Errorf("atr.toml:%d: unknown key %q", i+1, key)
		}
	}
	return cfg, nil
}

// parseValue accepts a double-quoted string (with \" \\ \n \t escapes), a
// single-quoted literal string (no escapes) or a bare true/false, optionally
// followed by a # comment. Both string forms follow standard TOML semantics.
func parseValue(s string) (strVal string, boolVal, isBool bool, err error) {
	if token, rest, ok := cutToken(s); ok && (token == "true" || token == "false") {
		if rest != "" && !strings.HasPrefix(rest, "#") {
			return "", false, false, fmt.Errorf("unexpected content after value: %q", rest)
		}
		return "", token == "true", true, nil
	}
	if len(s) < 2 || (s[0] != '"' && s[0] != '\'') {
		return "", false, false, fmt.Errorf("value must be a quoted string or true/false")
	}
	val, rest, err := cutString(s)
	if err != nil {
		return "", false, false, err
	}
	rest = strings.TrimSpace(rest)
	if rest != "" && !strings.HasPrefix(rest, "#") {
		return "", false, false, fmt.Errorf("unexpected content after string: %q", rest)
	}
	return val, false, false, nil
}

// cutString consumes the leading quoted string and returns the remainder.
// Double quotes are TOML basic strings (escapes processed), single quotes
// are TOML literal strings (no escapes).
func cutString(s string) (val, rest string, err error) {
	if s[0] == '\'' {
		end := strings.IndexByte(s[1:], '\'')
		if end < 0 {
			return "", "", fmt.Errorf("unclosed string")
		}
		return s[1 : end+1], s[end+2:], nil
	}
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '"':
			return b.String(), s[i+1:], nil
		case '\\':
			i++
			if i >= len(s) {
				return "", "", fmt.Errorf("unclosed string")
			}
			switch s[i] {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				return "", "", fmt.Errorf(`unsupported escape \%c`, s[i])
			}
		default:
			b.WriteByte(s[i])
		}
	}
	return "", "", fmt.Errorf("unclosed string")
}

// cutToken splits off the first whitespace-delimited token.
func cutToken(s string) (token, rest string, ok bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return "", "", false
	}
	return fields[0], strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), fields[0])), true
}
