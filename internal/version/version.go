package version

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/fx"
)

var (
	buildVersion string = "dev"
	buildDate    string = ""
	buildCommit  string = ""
)

// Set sets the version information
func Set(version, date, commit string) {
	buildVersion = version
	buildDate = date
	buildCommit = commit
}

// PrintVersion prints the version information
func PrintVersion() {
	fmt.Printf("Version: %s\nDate: %s\nCommit: %s\n", buildVersion, buildDate, buildCommit)
}

// Version contains app version information
type Version struct {
	Version string
	Date    string
	Commit  string
}

// NewVersion returns a new version struct
func NewVersion() *Version {
	return &Version{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	}
}

// String returns a string representation of the version
func (s *Version) String() string {
	return fmt.Sprintf("Version: %s\nDate: %s\nCommit: %s\n", s.Version, s.Date, s.Commit)
}

// RegisterVersionHooks registers the version hooks
func RegisterVersionHooks(lc fx.Lifecycle, logger *slog.Logger, version *Version) {

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("build info", "version", version.Version, "date", version.Date, "commit", version.Commit)
			return nil
		},
	})
}
