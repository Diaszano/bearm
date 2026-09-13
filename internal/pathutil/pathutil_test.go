package pathutil

import "testing"

func TestIsEqualOrDescendant(t *testing.T) {
	tests := []struct {
		name string
		path string
		root string
		want bool
	}{
		{"equal paths", "/foo", "/foo", true},
		{"child path", "/foo/bar", "/foo", true},
		{"deep child", "/foo/bar/baz", "/foo", true},
		{"not descendant prefix overlap", "/foobar", "/foo", false},
		{"sibling", "/bar", "/foo", false},
		{"root descendant", "/foo", "/", true},
		{"root equal", "/", "/", true},
		{"trailing slash root", "/foo/bar", "/foo/", true},
		{"uncleaned path", "/foo//bar", "/foo", true},
		{"uncleaned root", "/foo/bar", "/foo//", true},
		{"dot segments", "/foo/bar/../bar", "/foo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEqualOrDescendant(tt.path, tt.root)
			if got != tt.want {
				t.Errorf("IsEqualOrDescendant(%q, %q) = %v, want %v", tt.path, tt.root, got, tt.want)
			}
		})
	}
}
