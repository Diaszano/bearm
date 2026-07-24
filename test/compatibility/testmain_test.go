package compatibility

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var bearmBinary string

func TestMain(m *testing.M) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	buildDir, err := os.MkdirTemp("", "bearm-compat-build-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bearmBinary = filepath.Join(buildDir, "bearm")
	command := exec.Command("go", "build", "-trimpath", "-o", bearmBinary, "./cmd/bearm")
	command.Dir = root
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		_ = os.RemoveAll(buildDir)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(buildDir)
	os.Exit(code)
}
