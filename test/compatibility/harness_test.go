package compatibility

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type processResult struct {
	Code   int
	Stdout string
	Stderr string
	Tree   []TreeEntry
}

var resolvedRm string

func init() {
	if path, err := exec.LookPath("rm"); err == nil {
		resolvedRm = path
	} else {
		resolvedRm = "/bin/rm"
	}
}

func TestCoreCompatibility(t *testing.T) {
	for _, testCase := range CoreCases() {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			rmResult := runCase(t, resolvedRm, false, testCase)
			bearmResult := runCase(t, bearmBinary, true, testCase)

			if rmResult.Code != bearmResult.Code {
				t.Fatalf("exit code: rm=%d bearm=%d\nrm stderr=%q\nbearm stderr=%q",
					rmResult.Code, bearmResult.Code, rmResult.Stderr, bearmResult.Stderr)
			}
			if testCase.CompareOut && rmResult.Stdout != bearmResult.Stdout {
				t.Fatalf("stdout:\nrm=%q\nbearm=%q", rmResult.Stdout, bearmResult.Stdout)
			}
			if testCase.CompareErr && normalizeProgram(rmResult.Stderr) != normalizeProgram(bearmResult.Stderr) {
				t.Fatalf("stderr:\nrm=%q\nbearm=%q", rmResult.Stderr, bearmResult.Stderr)
			}
			if !reflect.DeepEqual(rmResult.Tree, bearmResult.Tree) {
				t.Fatalf("tree:\nrm=%#v\nbearm=%#v", rmResult.Tree, bearmResult.Tree)
			}
		})
	}
}

func runCase(t *testing.T, binary string, bearm bool, testCase Case) processResult {
	t.Helper()

	root := t.TempDir()
	if testCase.Setup != nil {
		if err := testCase.Setup(root); err != nil {
			t.Fatal(err)
		}
	}

	args := append([]string(nil), testCase.Args...)
	if bearm {
		args = append([]string{"rm"}, args...)
	}

	command := exec.Command(binary, args...)
	command.Dir = root
	command.Stdin = strings.NewReader(testCase.Stdin)
	command.Env = append(os.Environ(),
		"LC_ALL=C",
		"LANG=C",
		"BEARM_LANG=en",
		"BEARM_COMPAT="+hostProfile(),
		"BEARM_CONFIG_HOME="+filepath.Join(root, ".bearm-config"),
		"BEARM_STATE_HOME="+filepath.Join(root, ".bearm-state"),
		"BEARM_TRASH="+filepath.Join(root, ".bearm-trash"),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	code := 0
	if err := command.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %s: %v", binary, err)
		}
		code = exitErr.ExitCode()
	}

	tree, err := NormalizeTree(root)
	if err != nil {
		t.Fatal(err)
	}
	return processResult{
		Code: code, Stdout: stdout.String(), Stderr: stderr.String(), Tree: tree,
	}
}

func hostProfile() string {
	if runtime.GOOS == "darwin" {
		return "bsd"
	}
	return "gnu"
}

func normalizeProgram(value string) string {
	value = strings.ReplaceAll(value, bearmBinary, "rm")
	value = strings.ReplaceAll(value, resolvedRm, "rm")
	value = strings.ReplaceAll(value, "/bin/rm", "rm")
	return value
}
