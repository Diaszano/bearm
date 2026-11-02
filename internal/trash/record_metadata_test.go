package trash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestRecordMetadataPath(t *testing.T) {
	for _, tc := range []struct{ backend, base, suffix string }{
		{"linux-freedesktop", "/trash/files/a.1", ".trashinfo"},
		{"custom-trash", "/trash/files/a.1", ".json"},
		{"darwin-system-trash", "/home/.Trash/a.1", ".json"},
	} {
		t.Run(tc.backend, func(t *testing.T) {
			got, err := RecordMetadataPath(domain.TrashRecord{Backend: tc.backend, TrashedPath: tc.base})
			if err != nil {
				t.Fatal(err)
			}
			var want string
			switch tc.backend {
			case "linux-freedesktop":
				want = filepath.Join("/trash", "info", "a.1"+tc.suffix)
			case "custom-trash":
				want = filepath.Join("/trash", "info", "a.1"+tc.suffix)
			default:
				want = filepath.Join("/home/.Trash", ".bearm-info", "a.1"+tc.suffix)
			}
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	}
}

func TestRemoveRecordMetadata(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	files := filepath.Join(root, "files")
	info := filepath.Join(root, "info")
	if err := os.MkdirAll(files, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(info, 0700); err != nil {
		t.Fatal(err)
	}
	sidecar := filepath.Join(info, "item.json")
	if err := os.WriteFile(sidecar, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	rec := domain.TrashRecord{Backend: "custom-trash", TrashedPath: filepath.Join(files, "item")}
	if err := RemoveRecordMetadata(rec); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sidecar); !os.IsNotExist(err) {
		t.Fatalf("sidecar remains: %v", err)
	}
	if err := RemoveRecordMetadata(rec); err != nil {
		t.Fatalf("missing sidecar: %v", err)
	}
	if err := RemoveRecordMetadata(domain.TrashRecord{Backend: "unknown", TrashedPath: rec.TrashedPath}); err == nil {
		t.Fatal("unknown backend accepted")
	}
}

func TestRemoveRecordMetadataRejectsSymlinkDirectory(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	sidecar := filepath.Join(outside, "item.json")
	if err := os.WriteFile(sidecar, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "info")); err != nil {
		t.Fatal(err)
	}
	rec := domain.TrashRecord{Backend: "custom-trash", TrashedPath: filepath.Join(root, "files", "item")}
	if err := RemoveRecordMetadata(rec); err == nil {
		t.Fatal("symlink metadata directory accepted")
	}
	if _, err := os.Stat(sidecar); err != nil {
		t.Fatalf("outside sidecar removed: %v", err)
	}
}
