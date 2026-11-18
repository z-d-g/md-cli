package app

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/z-d-g/md-cli/v2/internal/utils"
)

type CLIArgs struct {
	PrintOnly bool
	Help      bool
	Version   bool
	Files     []string
}

// ParseCLIArgs parses argv. Flag errors are reported once, on stderr, and exit
// non-zero; help and version are returned so main controls output streams.
func ParseCLIArgs(args []string) CLIArgs {
	var printFlag, helpFlag, versionFlag bool
	fs := flag.NewFlagSet("md-cli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.BoolVar(&printFlag, "p", false, "Print rendered markdown to stdout")
	fs.BoolVar(&printFlag, "print", false, "Print rendered markdown to stdout")
	fs.BoolVar(&helpFlag, "h", false, "Show help")
	fs.BoolVar(&helpFlag, "help", false, "Show help")
	fs.BoolVar(&versionFlag, "v", false, "Show version")
	fs.BoolVar(&versionFlag, "version", false, "Show version")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return CLIArgs{Help: true}
		}
		fmt.Fprintf(os.Stderr, "md-cli: %v\n", err)
		PrintUsage(os.Stderr)
		os.Exit(2)
	}

	if helpFlag {
		return CLIArgs{Help: true}
	}
	if versionFlag {
		return CLIArgs{Version: true}
	}

	return CLIArgs{
		PrintOnly: printFlag,
		Files:     utils.FilterMarkdownFiles(fs.Args()),
	}
}

// PrintUsage writes the usage text to w.
func PrintUsage(w io.Writer) {
	fmt.Fprint(w, "Usage: md-cli [options] [file.md]\n\n")
	fmt.Fprint(w, "Options:\n")
	fmt.Fprint(w, "  -p, --print      Print rendered markdown to stdout\n")
	fmt.Fprint(w, "  -v, --version    Show version information\n")
	fmt.Fprint(w, "  -h, --help       Show help\n")
}
