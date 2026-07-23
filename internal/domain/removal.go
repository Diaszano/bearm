// Package domain contains Bearm's infrastructure-independent domain types.
package domain

import (
	"errors"
	"time"
)

// CompatibilityProfile selects the rm behavior Bearm emulates.
type CompatibilityProfile string

const (
	// ProfileGNU selects GNU coreutils-style parsing and diagnostics.
	ProfileGNU CompatibilityProfile = "gnu"
	// ProfileBSD selects BSD/macOS-style parsing and diagnostics.
	ProfileBSD CompatibilityProfile = "bsd"
	// ProfilePOSIX selects the portable POSIX subset.
	ProfilePOSIX CompatibilityProfile = "posix"
)

// InteractiveMode controls removal confirmation behavior.
type InteractiveMode string

const (
	// InteractiveDefault lets the compatibility profile choose its default.
	InteractiveDefault InteractiveMode = "default"
	// InteractiveNever disables confirmation.
	InteractiveNever InteractiveMode = "never"
	// InteractiveOnce prompts once for risky multi-target operations.
	InteractiveOnce InteractiveMode = "once"
	// InteractiveAlways prompts for every logical removal.
	InteractiveAlways InteractiveMode = "always"
)

// PreserveRootMode controls root and mount-root protection.
type PreserveRootMode string

const (
	// PreserveRootDefault protects the system root.
	PreserveRootDefault PreserveRootMode = "default"
	// PreserveRootNone disables compatibility-level root preservation.
	PreserveRootNone PreserveRootMode = "none"
	// PreserveRootAll protects command-line directories on distinct devices.
	PreserveRootAll PreserveRootMode = "all"
)

// RemoveOptions contains parsed compatibility-mode removal options.
type RemoveOptions struct {
	Force         bool
	Recursive     bool
	Directory     bool
	Verbose       bool
	Interactive   InteractiveMode
	PreserveRoot  PreserveRootMode
	OneFileSystem bool
}

// RemoveRequest is a compatibility-mode removal request.
type RemoveRequest struct {
	Profile  CompatibilityProfile
	Operands []string
	Options  RemoveOptions
}

// Validate validates profile-independent request invariants.
func (r RemoveRequest) Validate() error {
	switch r.Profile {
	case ProfileGNU, ProfileBSD, ProfilePOSIX:
	default:
		return errors.New("unsupported compatibility profile")
	}

	if len(r.Operands) == 0 && !r.Options.Force {
		return errors.New("missing operand")
	}

	return nil
}

// TargetKind classifies a planned filesystem target.
type TargetKind string

const (
	// TargetFile identifies a regular file.
	TargetFile TargetKind = "file"
	// TargetDir identifies a directory that is not a symlink.
	TargetDir TargetKind = "directory"
	// TargetSymlink identifies a symbolic link.
	TargetSymlink TargetKind = "symlink"
	// TargetOther identifies another filesystem object.
	TargetOther TargetKind = "other"
)

// PlannedTarget contains immutable target data captured during planning.
type PlannedTarget struct {
	InputPath    string
	AbsolutePath string
	Kind         TargetKind
	DeviceID     uint64
	RequiresWalk bool
}

// RemovalPlan is the validated set of targets for one operation.
type RemovalPlan struct {
	ID        string
	CreatedAt time.Time
	Request   RemoveRequest
	Targets   []PlannedTarget
}
