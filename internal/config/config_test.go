package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	cfg, err := parse(`
# comment
command = "python3 main.py"   # trailing comment
build = 'cabal build {task}'  # literal string (single quotes)
contest_template = "./ctpl"
task_template = "./ttpl"
select = true   # bool value
`, "/base")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Command != "python3 main.py" {
		t.Errorf("Command = %q", cfg.Command)
	}
	if cfg.Build != "cabal build {task}" {
		t.Errorf("Build = %q", cfg.Build)
	}
	if cfg.ContestTemplate != filepath.Join("/base", "ctpl") {
		t.Errorf("ContestTemplate = %q", cfg.ContestTemplate)
	}
	if cfg.TaskTemplate != filepath.Join("/base", "ttpl") {
		t.Errorf("TaskTemplate = %q", cfg.TaskTemplate)
	}
	if !cfg.Select {
		t.Error("Select should be true")
	}

	// standard TOML string forms: basic-string escapes and literal strings
	for src, want := range map[string]string{
		`command = "say \"hi\""`:        `say "hi"`,
		`command = "tab\there"`:         "tab\there",
		`command = "back\\slash"`:       `back\slash`,
		`command = 'no \n escapes # x'`: `no \n escapes # x`,
	} {
		cfg, err := parse(src, "/base")
		if err != nil || cfg.Command != want {
			t.Errorf("parse(%q): Command = %q, err = %v; want %q", src, cfg.Command, err, want)
		}
	}

	// outside the subset -> error, never a silent misread
	for _, bad := range []string{
		`command = 'unclosed`,
		`command = "bad \x escape"`,
		`command = 'a' trailing`,
		"command = unquoted",
		"[section]",
		`arr = ["a", "b"]`,
		`typo_key = "x"`,
		`template = "./tpl"`, // removed key (split into contest_template / task_template)
		`command = "a" trailing`,
		`command = "back\slash"`,
		`select = "true"`, // must be a bare bool
		`command = true`,  // must be a string
		`select = true extra`,
	} {
		if _, err := parse(bad, "/base"); err == nil {
			t.Errorf("parse(%q) should fail", bad)
		}
	}
}

func TestFindNearest(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "atr.toml"), []byte(`command = "outer"`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Find(sub) // found in an ancestor
	if err != nil || cfg.Command != "outer" {
		t.Fatalf("ancestor config: got %+v, %v", cfg, err)
	}

	if err := os.WriteFile(filepath.Join(sub, "atr.toml"), []byte(`command = "inner"`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _ = Find(sub) // nearest wins, no merging
	if cfg.Command != "inner" {
		t.Fatalf("nearest config: got %q", cfg.Command)
	}
}

func TestFindAbsent(t *testing.T) {
	cfg, err := Find(t.TempDir())
	if err != nil || cfg.Command != "./a.out" || cfg.ContestTemplate != "" || cfg.TaskTemplate != "" || cfg.Select {
		t.Fatalf("fallback config: got %+v, %v", cfg, err)
	}
}
