// Package buildinfo exposes immutable release metadata.
package buildinfo

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Info contains Bearm release metadata.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Current returns the metadata injected at build time.
func Current() Info {
	return Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

// String returns a stable human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("Bearm %s (%s, %s)", i.Version, i.Commit, i.Date)
}
