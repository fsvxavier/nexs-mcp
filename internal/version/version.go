// Package version provides version information for the nexs-mcp application.
package version

const (
	// VERSION is the current version of nexs-mcp
	VERSION = "1.1.0"

	// BuildDate is set during build time via ldflags
	BuildDate = "unknown"

	// GitCommit is set during build time via ldflags
	GitCommit = "unknown"
)

// Info returns version information as a struct
type Info struct {
	Version   string
	BuildDate string
	GitCommit string
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version:   VERSION,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
	}
}

// String returns a formatted version string
func String() string {
	return VERSION
}
