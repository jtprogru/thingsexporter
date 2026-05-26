package cli

import (
	"fmt"
	"runtime"

	"github.com/jtprogru/thingsexporter/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build version, commit, date, and toolchain",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(deps.Stdout,
				"thingsexporter %s\n  commit:    %s\n  built:     %s\n  built by:  %s\n  go:        %s\n  platform:  %s/%s\n",
				version.Version,
				orDash(version.Commit),
				orDash(version.Date),
				orDash(version.BuiltBy),
				runtime.Version(),
				runtime.GOOS, runtime.GOARCH,
			)
			return err
		},
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
