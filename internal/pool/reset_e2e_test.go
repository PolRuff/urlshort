package pool

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResetGeneratorE2E(t *testing.T) {
	// 1. Запускаем генератор
	cmd := exec.Command(
		"go",
		"run",
		"./cmd/reset",
	)

	cmd.Dir = projectRoot(t)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("reset generator failed:\n%s", out)
	}

	// 2. Проверяем поведение сгенерированного Reset()
	ts := &TestStruct{
		Value: 42,
		Name:  "test",
		Tags:  []string{"a", "b"},
		Settings: map[string]bool{
			"debug": true,
		},
		Nested: &TestStruct{
			Value: 10,
			Name:  "nested",
			Tags:  []string{"x"},
			Settings: map[string]bool{
				"enabled": true,
			},
		},
	}

	ts.Reset()

	if ts.Value != 0 {
		t.Errorf("Value not reset")
	}
	if ts.Name != "" {
		t.Errorf("Name not reset")
	}
	if len(ts.Tags) != 0 {
		t.Errorf("Tags not reset")
	}
	if len(ts.Settings) != 0 {
		t.Errorf("Settings not reset")
	}

	if ts.Nested == nil {
		t.Fatalf("Nested was nil after Reset")
	}
	if ts.Nested.Value != 0 {
		t.Errorf("Nested.Value not reset")
	}
	if ts.Nested.Name != "" {
		t.Errorf("Nested.Name not reset")
	}
	if len(ts.Nested.Tags) != 0 {
		t.Errorf("Nested.Tags not reset")
	}
	if len(ts.Nested.Settings) != 0 {
		t.Errorf("Nested.Settings not reset")
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller info")
	}

	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
