package config

import (
	"charm.land/lipgloss/v2"
)

// Theme holds hex color strings for all UI and markdown elements.
type Theme struct {
	TitleBg      string
	TitleFg      string
	LineNumberFg string
	CursorBg     string
	CursorFg     string

	H1Fg              string
	H2Fg              string
	H3Fg              string
	H4Fg              string
	H5Fg              string
	H6Fg              string
	BoldAttr          bool
	ItalicAttr        bool
	CodeSpanBg        string
	CodeSpanFg        string
	CodeBlockBg       string
	CodeBlockFg       string
	CodeFenceFg       string
	LinkFg            string
	LinkURLFg         string
	BulletFg          string
	BlockquoteFg      string
	ImageFg           string
	HRFg              string
	StrikethroughAttr bool
	TableBorderFg     string
	TableHeaderFg     string
	TableCellFg       string

	SelectionBg string
	SelectionFg string

	InfoFg        string
	HintFg        string
	UnsavedFg     string
	SavedFg       string
	HelpKeyFg     string
	HelpDescFg    string
	HelpSectionFg string
	ModalBorderFg string
}

// DefaultTheme returns the built-in dark theme.
func DefaultTheme() *Theme {
	return &Theme{
		// UI chrome
		TitleBg:      "#7D56F4", // blue-violet
		TitleFg:      "#FFFFFF", // white
		LineNumberFg: "#626262", // dim gray
		CursorBg:     "#D787D7", // soft purple
		CursorFg:     "#FFFFFF", // white

		// Markdown elements — gradient: orchid → lilac → teal → seafoam → steel → neutral
		H1Fg:              "#D75FD7", // soft orchid
		H2Fg:              "#AF87FF", // lilac
		H3Fg:              "#87AFAF", // muted teal
		H4Fg:              "#5FAFAF", // seafoam
		H5Fg:              "#5F87AF", // steel blue
		H6Fg:              "#8A8A8A", // neutral gray
		BoldAttr:          true,
		ItalicAttr:        true,
		CodeSpanBg:        "#3A3A3A", // dark gray
		CodeSpanFg:        "#AFD7D7", // light gray
		CodeBlockBg:       "#303030", // very dark gray
		CodeBlockFg:       "#D0D0D0", // off-white
		CodeFenceFg:       "#585858", // medium gray
		LinkFg:            "#5FD7FF", // bright cyan
		LinkURLFg:         "#808080", // medium gray
		BulletFg:          "#AF87FF", // lilac
		BlockquoteFg:      "#808080", // medium gray
		ImageFg:           "#7D56F4", // blue-violet
		HRFg:              "#585858", // medium gray
		StrikethroughAttr: true,
		TableBorderFg:     "#767676", // gray for borders
		TableHeaderFg:     "#AF87FF", // lilac for headers
		TableCellFg:       "#D0D0D0", // off-white for cells

		// Selection highlighting
		SelectionBg: "#7D56F4", // blue-violet background for selection
		SelectionFg: "#FFFFFF", // white text for selection

		// UI details
		InfoFg:        "#8A8A8A", // neutral gray
		HintFg:        "#808080", // medium gray
		UnsavedFg:     "#FFAF00", // orange
		SavedFg:       "#00AF5F", // green
		HelpKeyFg:     "#AF87FF", // lilac
		HelpDescFg:    "#D0D0D0", // off-white
		HelpSectionFg: "#767676", // gray
		ModalBorderFg: "#AF87FF", // lilac
	}
}

// ToEditorStyles converts Theme fields to lipgloss styles.
func (t *Theme) ToEditorStyles() EditorStyles {
	return EditorStyles{
		H1: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H1Fg)).
			Bold(t.BoldAttr),
		H2: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H2Fg)).
			Bold(t.BoldAttr),
		H3: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H3Fg)).
			Bold(t.BoldAttr),
		H4: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H4Fg)).
			Bold(t.BoldAttr),
		H5: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H5Fg)).
			Bold(t.BoldAttr),
		H6: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.H6Fg)).
			Bold(t.BoldAttr),
		Bold: lipgloss.NewStyle().
			Bold(t.BoldAttr),
		Italic: lipgloss.NewStyle().
			Italic(t.ItalicAttr),
		Underline: lipgloss.NewStyle().
			Underline(true),
		CodeSpan: lipgloss.NewStyle().
			Background(lipgloss.Color(t.CodeSpanBg)).
			Foreground(lipgloss.Color(t.CodeSpanFg)),
		CodeFence: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.CodeFenceFg)),
		CodeContent: lipgloss.NewStyle().
			Background(lipgloss.Color(t.CodeBlockBg)).
			Foreground(lipgloss.Color(t.CodeBlockFg)),
		Link: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.LinkFg)).
			Underline(true).
			UnderlineStyle(lipgloss.UnderlineCurly),
		LinkURL: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.LinkURLFg)),
		Bullet: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.BulletFg)),
		BlockQuote: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.BlockquoteFg)),
		Image: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.ImageFg)).
			Italic(t.ItalicAttr),
		HorizontalRule: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.HRFg)),
		Strikethrough: lipgloss.NewStyle().
			Strikethrough(t.StrikethroughAttr),
		TableBorder: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.TableBorderFg)),
		TableHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.TableHeaderFg)).
			Bold(true),
		TableCell: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.TableCellFg)),
		Selection: lipgloss.NewStyle().
			Background(lipgloss.Color(t.SelectionBg)).
			Foreground(lipgloss.Color(t.SelectionFg)),
		Cursor: lipgloss.NewStyle().
			Background(lipgloss.Color(t.CursorBg)).
			Foreground(lipgloss.Color(t.CursorFg)),
		LineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.LineNumberFg)),
	}
}
