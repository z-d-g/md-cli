package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/z-d-g/md-cli/internal/utils"
)

type CLIArgs struct {
	PrintOnly bool
	Help      bool
	Files     []string
}

func ParseCLIArgs(args []string) CLIArgs {
	var printFlag, helpFlag bool
	fs := flag.NewFlagSet("md-cli", flag.ContinueOnError)
	fs.BoolVar(&printFlag, "p", false, "Print rendered markdown to stdout")
	fs.BoolVar(&printFlag, "print", false, "Print rendered markdown to stdout")
	fs.BoolVar(&helpFlag, "h", false, "Show help")
	fs.BoolVar(&helpFlag, "help", false, "Show help")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return CLIArgs{Help: true}
		}
		PrintUsage()
		os.Exit(1)
	}

	return CLIArgs{
		PrintOnly: printFlag,
		Help:      helpFlag,
		Files:     utils.FilterMarkdownFiles(fs.Args()),
	}
}

func PrintUsage() {
	fmt.Println("Usage: md-cli [options] file.md")
	fmt.Println("Options:")
	fmt.Println("  -p, --print    Print rendered markdown to stdout")
	fmt.Println("  -h, --help     Show help")
}
