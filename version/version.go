// Package version provides compile-time version information.
package version

import (
	"runtime"
	"runtime/debug"
	"strconv"
)

// GitTreeState represents the state of the Git repository's working tree
// at the time of compilation. It provides a type-safe way to indicate whether
// the source tree was clean, modified, or in an unknown state.
type GitTreeState string

func (g GitTreeState) String() string {
	return string(g)
}

// These constants define all possible values for the GitTreeState contained in Info.
const (
	GitTreeStateModified GitTreeState = "modified"
	GitTreeStateClean    GitTreeState = "clean"
	GitTreeStateUnknown  GitTreeState = "unknown"
)

// Info carries the information gathered and returned by GetVersion().
type Info struct {
	Version       string
	GitCommit     string
	GitCommitTime string
	GitTreeState  GitTreeState
	GoVersion     string
}

// Get returns the version of the application derived directly from the runtime's debug information.
func Get() Info {
	var res Info
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		res.Version = "unknown"
	}

	res.Version = bi.Main.Version
	for _, bs := range bi.Settings {
		switch bs.Key {
		case "vcs.revision":
			res.GitCommit = bs.Value
		case "vcs.modified":
			mod, err := strconv.ParseBool(bs.Value)
			if err != nil {
				res.GitTreeState = GitTreeStateUnknown
				continue
			}
			if mod {
				res.GitTreeState = GitTreeStateModified
				continue
			}
			res.GitTreeState = GitTreeStateClean
		case "vcs.time":
			res.GitCommitTime = bs.Value
		default:
			// we don't care about other settings for now
		}
	}

	res.GoVersion = runtime.Version()

	return res
}
