// Package compatibility differentially tests Bearm against the host rm.
package compatibility

import (
	"os"
	"path/filepath"
)

// Case is one host rm differential scenario.
type Case struct {
	Name       string
	Args       []string
	Stdin      string
	Setup      func(root string) error
	CompareOut bool
	CompareErr bool
}

// CoreCases returns safe isolated compatibility scenarios.
func CoreCases() []Case {
	return []Case{
		{
			Name:       "missing operand",
			CompareErr: true,
		},
		{
			Name:       "force missing operand",
			Args:       []string{"-f"},
			CompareOut: true,
			CompareErr: true,
		},
		{
			Name:       "remove regular file",
			Args:       []string{"file.txt"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "remove missing file",
			Args:       []string{"missing.txt"},
			CompareOut: true,
			CompareErr: true,
		},
		{
			Name:       "recursive directory",
			Args:       []string{"-rf", "directory"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				if err := os.MkdirAll(filepath.Join(root, "directory", "sub"), 0o700); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(root, "directory", "sub", "file"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "directory without recursive",
			Args:       []string{"directory"},
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.Mkdir(filepath.Join(root, "directory"), 0o700)
			},
		},
		{
			Name:       "double dash filename",
			Args:       []string{"--", "-rf"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "-rf"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "force then interactive declined",
			Args:       []string{"-f", "-i", "file.txt"},
			Stdin:      "n\n",
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "interactive then force",
			Args:       []string{"-i", "-f", "file.txt"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "filename with spaces and unicode",
			Args:       []string{"ação com espaço.txt"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "ação com espaço.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "dangling symlink",
			Args:       []string{"dangling"},
			CompareOut: true,
			CompareErr: true,
			Setup: func(root string) error {
				return os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "dangling"))
			},
		},
		{
			Name:       "trailing slash on file",
			Args:       []string{"file.txt/"},
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.WriteFile(filepath.Join(root, "file.txt"), []byte("data"), 0o600)
			},
		},
		{
			Name:       "empty directory with d",
			Args:       []string{"-d", "empty"},
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				return os.Mkdir(filepath.Join(root, "empty"), 0o700)
			},
		},
		{
			Name:       "interactive once four operands declined",
			Args:       []string{"-I", "a", "b", "c", "d"},
			Stdin:      "n\n",
			CompareOut: true,
			CompareErr: false,
			Setup: func(root string) error {
				for _, name := range []string{"a", "b", "c", "d"} {
					if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o600); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}
