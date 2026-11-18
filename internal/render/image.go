package render

import (
	"strings"

	"github.com/z-d-g/md-cli/v2/internal/config"
)

// renderImageAlt renders an image with its alt text.
func renderImageAlt(alt string, styles *config.EditorStyles) string {
	var b strings.Builder
	b.WriteString(styles.Image.Render("⊞ "))
	if alt != "" {
		b.WriteString(styles.Image.Render(alt))
	} else {
		b.WriteString(styles.Image.Render("[image]"))
	}
	return b.String()
}
