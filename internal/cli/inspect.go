package cli

import (
	"context"
	encjson "encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type inspectFlags struct {
	db    string
	quiet bool
}

func newInspectCmd(deps Deps) *cobra.Command {
	var f inspectFlags
	cmd := &cobra.Command{
		Use:           "inspect",
		Short:         "Print counts and databaseVersion of Things 3 DB without exporting",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInspect(cmd.Context(), deps, f)
		},
	}
	cmd.Flags().StringVar(&f.db, "db", "", "Path to Things 3 main.sqlite (default: auto-discover on macOS)")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "Suppress informational messages in stderr")
	return cmd
}

func runInspect(ctx context.Context, deps Deps, f inspectFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	path, err := resolveDBPath(deps, f.db, f.quiet)
	if err != nil {
		return err
	}
	repo, err := deps.OpenRepo(path)
	if err != nil {
		return exit2(fmt.Errorf("cannot open db: %w", err))
	}
	defer func() { _ = repo.Close() }()

	counts, err := repo.ReadCounts(ctx)
	if err != nil {
		return exit2(fmt.Errorf("cannot read counts: %w", err))
	}
	version, _ := repo.DatabaseVersion(ctx)

	if version != nil && !f.quiet && !versionSupported(*version, deps.SupportedDBVersions) {
		_, _ = fmt.Fprintf(deps.Stderr, "warning: unsupported Things 3 databaseVersion=%d, output may be incomplete\n", *version)
	}

	payload := map[string]any{
		"path":            path,
		"databaseVersion": version,
		"counts":          counts,
	}
	enc := encjson.NewEncoder(deps.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return exit2(fmt.Errorf("write: %w", err))
	}
	return nil
}
