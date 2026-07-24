package planner_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/planner"
)

func TestWalkDepthFirstDoesNotFollowSymlinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dir", "file"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "dir"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	got, err := planner.WalkDepthFirst(root, 0, false)
	if err != nil {
		t.Fatalf("WalkDepthFirst() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "dir", "file"),
		filepath.Join(root, "dir"),
		filepath.Join(root, "link"),
		root,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestWalkDepthFirstMultiLevelOrdering(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dirA := filepath.Join(root, "dirA")
	subdirB := filepath.Join(dirA, "subdirB")
	if err := os.MkdirAll(subdirB, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirA, "fileA"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdirB, "fileB"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := planner.WalkDepthFirst(root, 0, false)
	if err != nil {
		t.Fatalf("WalkDepthFirst() error = %v", err)
	}

	want := []string{
		filepath.Join(dirA, "fileA"),
		filepath.Join(subdirB, "fileB"),
		subdirB,
		dirA,
		root,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestWalkDepthFirstOneFileSystemSameDevice(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := planner.WalkDepthFirst(root, 0, true)
	if err != nil {
		t.Fatalf("WalkDepthFirst() error = %v", err)
	}

	want := []string{file, root}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}
