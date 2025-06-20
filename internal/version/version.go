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

// Set - устанавливает значения версии, даты и коммита
func Set(version, date, commit string) {
	buildVersion = version
	buildDate = date
	buildCommit = commit
}

type Version struct {
	Version string
	Date    string
	Commit  string
}

func NewVersion() *Version {
	return &Version{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	}
}

func (s *Version) String() string {
	return fmt.Sprintf("Version: %s\nDate: %s\nCommit: %s\n", s.Version, s.Date, s.Commit)
}

func RegisterVersionHooks(lc fx.Lifecycle, log *slog.Logger, version *Version) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Build info", "version", version.Version, "date", version.Date, "commit", version.Commit)
			return nil
		},
	})
}
