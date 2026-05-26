// Command thingsexporter — CLI для экспорта БД Things 3 в JSON или Markdown.
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/jtprogru/thingsexporter/internal/cli"
)

func main() {
	exitCode := 0
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "thingsexporter: panic: %v\n%s\n", r, debug.Stack())
			os.Exit(1)
		}
		os.Exit(exitCode)
	}()

	deps := cli.DefaultDeps()
	if err := cli.Execute(deps); err != nil {
		exitCode = cli.AsExitCode(err)
	}
}
