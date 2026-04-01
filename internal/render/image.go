package render

import "strings"

// renderImageIcon renders a simple image placeholder icon.
func renderImageIcon(styles *styleCache) string {
	return styles.imageFunc("⊞")
}

// renderImageAlt renders an image with its alt text.
func renderImageAlt(alt string, styles *styleCache) string {
	var b strings.Builder
	b.WriteString(styles.imageFunc("⊞ "))
	if alt != "" {
		b.WriteString(styles.imageFunc(alt))
	} else {
		b.WriteString(styles.imageFunc("[image]"))
	}
	return b.String()
}
