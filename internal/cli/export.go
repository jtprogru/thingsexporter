package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jtprogru/thingsexporter/internal/export"
	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/spf13/cobra"
)

type exportFlags struct {
	db      string
	out     string
	format  string
	include string
	indent  int
	noBlobs bool
	quiet   bool
}

func defaultExportFlags() exportFlags {
	return exportFlags{
		out:     "-",
		format:  "json",
		include: "all",
		indent:  2,
	}
}

func newExportCmd(deps Deps) *cobra.Command {
	f := defaultExportFlags()
	cmd := &cobra.Command{
		Use:           "export",
		Short:         "Export Things 3 database to JSON or Markdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(cmd.Context(), deps, f)
		},
	}
	addExportFlags(cmd, &f)
	return cmd
}

func addExportFlags(cmd *cobra.Command, f *exportFlags) {
	cmd.Flags().StringVar(&f.db, "db", f.db, "Path to Things 3 main.sqlite (default: auto-discover on macOS)")
	cmd.Flags().StringVar(&f.out, "out", f.out, "Output path ('-' for stdout)")
	cmd.Flags().StringVar(&f.format, "format", f.format, "Output format: json | markdown")
	cmd.Flags().StringVar(&f.include, "include", f.include, "Content preset: all | tasks | tasks+tags | tasks+projects")
	cmd.Flags().IntVar(&f.indent, "indent", f.indent, "JSON indent (0 = compact)")
	cmd.Flags().BoolVar(&f.noBlobs, "no-blobs", f.noBlobs, "Strip BLOB fields from output (default: hex-encoded)")
	cmd.Flags().BoolVar(&f.quiet, "quiet", f.quiet, "Suppress summary in stderr")
}

func runExport(ctx context.Context, deps Deps, f exportFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	path, err := resolveDBPath(deps, f.db, f.quiet)
	if err != nil {
		return err
	}
	if f.indent < 0 {
		return exit2(errors.New("--indent must be >= 0"))
	}
	writer, err := deps.Writers.Lookup(f.format)
	if err != nil {
		return exit2(err)
	}
	p, err := deps.Presets.Lookup(f.include)
	if err != nil {
		return exit2(err)
	}

	repo, err := deps.OpenRepo(path)
	if err != nil {
		return exit2(fmt.Errorf("cannot open db: %w", err))
	}
	defer func() { _ = repo.Close() }()

	if v, verr := repo.DatabaseVersion(ctx); verr == nil && v != nil && !f.quiet {
		if !versionSupported(*v, deps.SupportedDBVersions) {
			_, _ = fmt.Fprintf(deps.Stderr, "warning: unsupported Things 3 databaseVersion=%d, output may be incomplete\n", *v)
		}
	}

	raw, err := repo.ReadAll(ctx)
	if err != nil {
		return exit2(fmt.Errorf("cannot read db: %w", err))
	}

	exported := things.Build(raw, things.BuildOptions{
		Source:     path,
		ExportedAt: deps.Clock().UTC(),
		NoBlobs:    f.noBlobs,
	})
	exported = p.Apply(exported)

	out, closeOut, err := openOutput(f.out, deps.Stdout)
	if err != nil {
		return exit2(err)
	}
	defer closeOut()

	if werr := writer.Write(out, exported, export.Options{Indent: f.indent}); werr != nil {
		return exit2(fmt.Errorf("write: %w", werr))
	}

	if !f.quiet {
		printSummary(deps.Stderr, f.out, exported.Meta.Counts)
	}
	return nil
}

func resolveDBPath(deps Deps, fromFlag string, quiet bool) (string, error) {
	if fromFlag != "" {
		return fromFlag, nil
	}
	if deps.DiscoverDB != nil {
		if p, ok := deps.DiscoverDB(); ok {
			if !quiet {
				_, _ = fmt.Fprintf(deps.Stderr, "using DB: %s\n", p)
			}
			return p, nil
		}
	}
	return "", exit2(errors.New("--db is required (no Things 3 database found at default path)"))
}

func openOutput(path string, stdout io.Writer) (io.Writer, func(), error) {
	if path == "" || path == "-" {
		return stdout, func() {}, nil
	}
	//nolint:gosec // path is user-supplied output target by design
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot write to %s: %w", path, err)
	}
	return f, func() { _ = f.Close() }, nil
}

func versionSupported(v int, supported []int) bool {
	for _, s := range supported {
		if s == v {
			return true
		}
	}
	return false
}

func printSummary(w io.Writer, outPath string, c things.Counts) {
	target := outPath
	if target == "" || target == "-" {
		target = "stdout"
	}
	_, _ = fmt.Fprintf(w, "OK -> %s\n", target)
	printCount(w, "areas", c.Areas)
	printCount(w, "tags", c.Tags)
	printCount(w, "tasks", c.Tasks)
	printCount(w, "checklistItems", c.ChecklistItems)
	printCount(w, "contacts", c.Contacts)
	printCount(w, "tombstones", c.Tombstones)
	printCount(w, "taskTagLinks", c.TaskTagLinks)
	printCount(w, "areaTagLinks", c.AreaTagLinks)
}

func printCount(w io.Writer, name string, v *int) {
	if v == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "  %s: %d\n", name, *v)
}
