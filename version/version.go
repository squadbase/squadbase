package version

import (
	"fmt"
	"strings"
)

const fallbackVersion = "dev"

var Version = fallbackVersion

var BuildTime = "unknown"

var GitCommit = "unknown"

func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built at: %s)",
		Version, GitCommit, BuildTime)
}

func IsDevVersion() bool {
	return Version == fallbackVersion || strings.Contains(Version, "dev")
}

func IsPreRelease() bool {
	return strings.Contains(Version, "alpha") ||
		strings.Contains(Version, "beta") ||
		strings.Contains(Version, "rc")
}
