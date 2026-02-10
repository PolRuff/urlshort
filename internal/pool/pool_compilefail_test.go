package pool

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCompileTimeConstraint(t *testing.T) {
	cmd := exec.Command(
		"go",
		"test",
		"./compilefail",
		"-tags=compilefail",
	)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected compilation to fail, but it succeeded")
	}

	if !strings.Contains(string(out), "does not satisfy") {
		t.Fatalf("unexpected error:\n%s", out)
	}

}
