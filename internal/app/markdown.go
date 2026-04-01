package app

import (
	"fmt"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/render"
	"github.com/z-d-g/md-cli/internal/utils"
)

func HandlePrintMode(files []string, cfg *config.Config) {
	r := render.NewLipglossRenderer(&cfg.EditorStyles)
	pr := render.NewPrintRenderer(r)

	for _, f := range files {
		content, err := utils.ReadFile(f)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", f, err)
			continue
		}
		fmt.Printf("--- %s ---\n%s\n", f, pr.RenderDocument(string(content)))
	}
}

func HandlePrintContent(content string, cfg *config.Config) {
	r := render.NewLipglossRenderer(&cfg.EditorStyles)
	pr := render.NewPrintRenderer(r)
	fmt.Print(pr.RenderDocument(content))
}
