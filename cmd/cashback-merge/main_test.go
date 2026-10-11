package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	config := "../../campaign.json"
	args := []string{"--config", config, "--seed", "341124"}
	var first, second, stderr bytes.Buffer
	if err := run(args, &first, &stderr); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"--config", config}, &second, &stderr); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() || strings.Count(first.String(), "%") != 2 ||
		!strings.Contains(first.String(), "Moves: 0/100 | Score: 0") {
		t.Fatalf("incorrect or non-repeatable initial field: %q", first.String())
	}
}

func TestRunRejectsInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`{"movesLimit": 0}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--config", path}, &stdout, &stderr); err == nil {
		t.Fatal("invalid config must fail")
	}
	if stdout.Len() != 0 {
		t.Fatal("invalid config must not start a game")
	}
}
