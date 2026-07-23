package cli

import (
	"errors"
	"strconv"
)

// NativeCommand identifies a Bearm-native command.
type NativeCommand string

const (
	CommandVersion NativeCommand = "version"
	CommandList    NativeCommand = "list"
	CommandRestore NativeCommand = "restore"
	CommandPurge   NativeCommand = "purge"
	CommandDoctor  NativeCommand = "doctor"
	CommandConfig  NativeCommand = "config"
)

// NativeRequest contains parsed native command arguments.
type NativeRequest struct {
	Command   NativeCommand
	ItemIDs   []string
	Operation string
	Last      bool
	Limit     int
	JSON      bool
	Yes       bool
	ConfigOp  string
}

// ParseNative parses Bearm-native command arguments.
func ParseNative(args []string) (NativeRequest, error) {
	if len(args) == 0 {
		return NativeRequest{}, errors.New("missing native command")
	}

	request := NativeRequest{Command: NativeCommand(args[0]), Limit: 50}
	switch request.Command {
	case CommandVersion:
		if len(args) != 1 {
			return NativeRequest{}, errors.New("version accepts no arguments")
		}
		return request, nil
	case CommandList:
		return parseNativeList(request, args[1:])
	case CommandRestore:
		return parseNativeSelection(request, args[1:], false)
	case CommandPurge:
		return parseNativeSelection(request, args[1:], true)
	case CommandDoctor:
		return parseNativeDoctor(request, args[1:])
	case CommandConfig:
		if len(args) != 2 || (args[1] != "path" && args[1] != "check") {
			return NativeRequest{}, errors.New("config requires path or check")
		}
		request.ConfigOp = args[1]
		return request, nil
	default:
		return NativeRequest{}, errors.New("unknown native command")
	}
}

func parseNativeList(request NativeRequest, args []string) (NativeRequest, error) {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--json":
			request.JSON = true
		case "--operation":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--operation requires a value")
			}
			request.Operation = args[index]
		case "--limit":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--limit requires a value")
			}
			value, err := strconv.Atoi(args[index])
			if err != nil || value < 1 {
				return NativeRequest{}, errors.New("--limit must be a positive integer")
			}
			request.Limit = value
		default:
			return NativeRequest{}, errors.New("unsupported list option")
		}
	}

	return request, nil
}

func parseNativeSelection(request NativeRequest, args []string, allowYes bool) (NativeRequest, error) {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--last":
			request.Last = true
		case "--operation":
			index++
			if index >= len(args) {
				return NativeRequest{}, errors.New("--operation requires a value")
			}
			request.Operation = args[index]
		case "--yes":
			if !allowYes {
				return NativeRequest{}, errors.New("--yes is valid only for purge")
			}
			request.Yes = true
		default:
			if len(args[index]) > 0 && args[index][0] == '-' {
				return NativeRequest{}, errors.New("unsupported selection option")
			}
			request.ItemIDs = append(request.ItemIDs, args[index])
		}
	}

	selectors := 0
	if request.Last {
		selectors++
	}
	if request.Operation != "" {
		selectors++
	}
	if len(request.ItemIDs) > 0 {
		selectors++
	}
	if selectors != 1 {
		return NativeRequest{}, errors.New("select exactly one of item IDs, --operation, or --last")
	}

	return request, nil
}

func parseNativeDoctor(request NativeRequest, args []string) (NativeRequest, error) {
	if len(args) == 0 {
		return request, nil
	}
	if len(args) == 1 && args[0] == "--json" {
		request.JSON = true
		return request, nil
	}
	return NativeRequest{}, errors.New("doctor accepts only --json")
}
