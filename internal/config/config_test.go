package config

import "testing"

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme == nil {
		t.Fatal("DefaultTheme() returned nil")
	}
	if theme.TitleBg == "" {
		t.Error("TitleBg is empty")
	}
	if theme.H1Fg == "" {
		t.Error("H1Fg is empty")
	}
}

func TestToEditorStyles(t *testing.T) {
	theme := DefaultTheme()
	styles := theme.ToEditorStyles()

	// Verify styles are populated (non-zero values for foreground)
	if styles.H1.GetForeground() == nil {
		t.Error("H1 foreground is nil")
	}
	if styles.Bold.GetBold() != theme.BoldAttr {
		t.Errorf("Bold attr = %v, want %v", styles.Bold.GetBold(), theme.BoldAttr)
	}
	if styles.Italic.GetItalic() != theme.ItalicAttr {
		t.Errorf("Italic attr = %v, want %v", styles.Italic.GetItalic(), theme.ItalicAttr)
	}
	if styles.CodeSpan.GetBackground() == nil {
		t.Error("CodeSpan background is nil")
	}
	if styles.Link.GetForeground() == nil {
		t.Error("Link foreground is nil")
	}
	if !styles.Strikethrough.GetStrikethrough() {
		t.Error("Strikethrough should be true")
	}
}

func TestBuildConfig(t *testing.T) {
	theme := DefaultTheme()

	cfg := buildConfig(theme)
	if cfg == nil {
		t.Fatal("buildConfig() returned nil")
	}
	if cfg.TitleStyle.GetForeground() == nil {
		t.Error("TitleStyle foreground is nil")
	}
}

func TestLoadConfig(t *testing.T) {
	cfg := LoadConfig()
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil")
	}
	if cfg.EditorStyles.H1.GetForeground() == nil {
		t.Error("H1 foreground is nil")
	}
}

func TestConfigStylesNotEmpty(t *testing.T) {
	cfg := LoadConfig()

	styles := []struct {
		name  string
		style interface{ GetForeground() any }
	}{
		// Just verify the config is non-nil and usable
	}
	_ = styles
	_ = cfg

	// Verify all editor styles have non-nil foregrounds
	es := cfg.EditorStyles
	if es.H2.GetForeground() == nil {
		t.Error("H2 foreground nil")
	}
	if es.H3.GetForeground() == nil {
		t.Error("H3 foreground nil")
	}
	if es.CodeContent.GetForeground() == nil {
		t.Error("CodeContent foreground nil")
	}
	if es.Bullet.GetForeground() == nil {
		t.Error("Bullet foreground nil")
	}
	if es.BlockQuote.GetForeground() == nil {
		t.Error("BlockQuote foreground nil")
	}
	if es.TableBorder.GetForeground() == nil {
		t.Error("TableBorder foreground nil")
	}
	if es.Selection.GetForeground() == nil {
		t.Error("Selection foreground nil")
	}
	if es.Cursor.GetForeground() == nil {
		t.Error("Cursor foreground nil")
	}
	if es.LineNumber.GetForeground() == nil {
		t.Error("LineNumber foreground nil")
	}
}
