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
template = "./tpl"
`, "/base")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Command != "python3 main.py" {
		t.Errorf("Command = %q", cfg.Command)
	}
	if cfg.Template != filepath.Join("/base", "tpl") {
		t.Errorf("Template = %q", cfg.Template)
	}

	// outside the subset -> error, never a silent misread
	for _, bad := range []string{
		"command = unquoted",
		"[section]",
		`arr = ["a", "b"]`,
		`typo_key = "x"`,
		`command = "a" trailing`,
		`command = "back\slash"`,
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
	if err != nil || cfg.Command != "./a.out" || cfg.Template != "" {
		t.Fatalf("fallback config: got %+v, %v", cfg, err)
	}
}
