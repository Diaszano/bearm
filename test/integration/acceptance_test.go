package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemoveListRestoreRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping subprocess acceptance test")
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	binary := filepath.Join(root, "bearm")
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/bearm")
	build.Dir = repositoryRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build error = %v\n%s", err, output)
	}

	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work")
	configRoot := filepath.Join(root, "config")
	stateRoot := filepath.Join(root, "state")
	trashRoot := filepath.Join(root, "trash")
	for _, directory := range []string{home, work, configRoot, stateRoot, trashRoot} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	source := filepath.Join(work, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(),
		"HOME="+home,
		"BEARM_CONFIG_HOME="+configRoot,
		"BEARM_STATE_HOME="+stateRoot,
		"BEARM_TRASH="+trashRoot,
		"BEARM_COMPAT="+map[bool]string{true: "bsd", false: "gnu"}[runtime.GOOS == "darwin"],
	)

	run := func(args ...string) string {
		command := exec.Command(binary, args...)
		command.Env = env
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%s error = %v\n%s", strings.Join(args, " "), err, output)
		}
		return string(output)
	}

	run("rm", source)
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	if output := run("list", "--json"); !strings.Contains(output, source) {
		t.Fatalf("list output = %q", output)
	}
	run("restore", "--last")
	if got, err := os.ReadFile(source); err != nil || string(got) != "data" {
		t.Fatalf("restored data = %q, error = %v", got, err)
	}
}
