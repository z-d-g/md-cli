package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/z-d-g/md-cli/v2/internal/config"
	"github.com/z-d-g/md-cli/v2/internal/render"
	"github.com/z-d-g/md-cli/v2/internal/utils"
	"golang.org/x/term"
)

// printWidth is the render width for non-interactive output. When stdout is a
// terminal we use its width, otherwise we fall back to 80 columns.
const printWidth = 80

func stdoutWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return printWidth
}

// HandlePrintMode renders each file to w. Read failures go to stderr and make
// the function return an error, so pipelines observe a non-zero exit status.
func HandlePrintMode(w io.Writer, files []string, cfg *config.Config) error {
	r := render.NewLipglossRenderer(&cfg.EditorStyles, stdoutWidth())
	pr := render.NewPrintRenderer(r)

	var failed bool
	for _, f := range files {
		content, err := utils.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "md-cli: %s: %v\n", f, err)
			failed = true
			continue
		}
		fmt.Fprintf(w, "--- %s ---\n%s\n", f, pr.RenderDocument(string(content)))
	}
	if failed {
		return fmt.Errorf("one or more files could not be read")
	}
	return nil
}

// HandlePrintContent renders markdown read from stdin to w.
func HandlePrintContent(w io.Writer, content string, cfg *config.Config) {
	r := render.NewLipglossRenderer(&cfg.EditorStyles, stdoutWidth())
	lines := strings.Split(content, "\n")
	r.SetDocument(func() []string { return lines })
	pr := render.NewPrintRenderer(r)
	fmt.Fprint(w, pr.RenderDocument(content))
}
