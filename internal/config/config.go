package config

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Config struct {
	TitleStyle       lipgloss.Style
	InfoStyle        lipgloss.Style
	HintStyle        lipgloss.Style
	UnsavedStyle     lipgloss.Style
	SavedStyle       lipgloss.Style
	HelpKeyStyle     lipgloss.Style
	HelpDescStyle    lipgloss.Style
	HelpSectionStyle lipgloss.Style
	ModalBorderStyle lipgloss.Style
	EditorStyles     EditorStyles
}

type EditorStyles struct {
	H1             lipgloss.Style
	H2             lipgloss.Style
	H3             lipgloss.Style
	H4             lipgloss.Style
	H5             lipgloss.Style
	H6             lipgloss.Style
	Bold           lipgloss.Style
	Italic         lipgloss.Style
	CodeSpan       lipgloss.Style
	CodeFence      lipgloss.Style
	CodeContent    lipgloss.Style
	Link           lipgloss.Style
	LinkURL        lipgloss.Style
	Bullet         lipgloss.Style
	BlockQuote     lipgloss.Style
	Image          lipgloss.Style
	HorizontalRule lipgloss.Style
	Strikethrough  lipgloss.Style
	Underline      lipgloss.Style
	TableBorder    lipgloss.Style
	TableHeader    lipgloss.Style
	TableCell      lipgloss.Style
	Selection      lipgloss.Style
	Cursor         lipgloss.Style
	LineNumber     lipgloss.Style
}

func buildConfig(theme *Theme, darkBG bool) *Config {
	fg := func(light, dark string) color.Color {
		if darkBG {
			return lipgloss.Color(dark)
		}
		return lipgloss.Color(light)
	}

	return &Config{
		TitleStyle: lipgloss.NewStyle().
			Background(fg(theme.TitleBg, theme.TitleBg)).
			Foreground(fg(theme.TitleFg, theme.TitleFg)).
			Padding(0, 1),

		InfoStyle: lipgloss.NewStyle().
			Foreground(fg(theme.InfoFg, theme.InfoFg)),

		HintStyle: lipgloss.NewStyle().
			Foreground(fg(theme.HintFg, theme.HintFg)).
			Italic(true),

		UnsavedStyle: lipgloss.NewStyle().
			Foreground(fg(theme.UnsavedFg, theme.UnsavedFg)).
			Bold(true),

		SavedStyle: lipgloss.NewStyle().
			Foreground(fg(theme.SavedFg, theme.SavedFg)).
			Bold(true),

		HelpKeyStyle: lipgloss.NewStyle().
			Foreground(fg(theme.HelpKeyFg, theme.HelpKeyFg)).
			Bold(true),

		HelpDescStyle: lipgloss.NewStyle().
			Foreground(fg(theme.HelpDescFg, theme.HelpDescFg)),

		HelpSectionStyle: lipgloss.NewStyle().
			Foreground(fg(theme.HelpSectionFg, theme.HelpSectionFg)).
			Bold(true),

		ModalBorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(fg(theme.ModalBorderFg, theme.ModalBorderFg)).
			Padding(0, 1),

		EditorStyles: theme.ToEditorStyles(),
	}
}

func LoadConfig() *Config {
	return buildConfig(DefaultTheme(), true)
}

func LoadConfigAdaptive(hasDarkBG bool) *Config {
	return buildConfig(DefaultTheme(), hasDarkBG)
}
