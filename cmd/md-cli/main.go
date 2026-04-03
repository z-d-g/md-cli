package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/z-d-g/md-cli/internal/app"
	"github.com/z-d-g/md-cli/internal/config"

	tea "charm.land/bubbletea/v2"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cliArgs := app.ParseCLIArgs(os.Args[1:])

	if cliArgs.Version {
		printVersion()
		return
	}

	if cliArgs.Help {
		app.PrintUsage()
		return
	}

	pipedContent, hasStdin := readStdin()

	cfg := config.LoadConfig()

	if cliArgs.PrintOnly {
		if hasStdin {
			app.HandlePrintContent(pipedContent, cfg)
		} else {
			app.HandlePrintMode(cliArgs.Files, cfg)
		}
		return
	}

	if hasStdin {
		fmt.Fprintln(os.Stderr, "Interactive mode requires --print when reading from stdin.")
		fmt.Fprintln(os.Stderr, "Usage: cat file.md | md-cli --print")
		os.Exit(1)
	}

	if len(cliArgs.Files) == 0 {
		app.PrintUsage()
		os.Exit(1)
	}

	p := tea.NewProgram(app.NewModel(cliArgs.Files[0], cfg))
	if _, err := p.Run(); err != nil {
		slog.Error("run", "err", err)
		os.Exit(1)
	}
}

func readStdin() (string, bool) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", false
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", false
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		return "", false
	}
	return string(data), true
}

func printVersion() {
	fmt.Printf("md-cli %s (commit: %s, built: %s)\n",
		version, commit, date)
}
