package app

import (
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/z-d-g/md-cli/internal/config"
	"github.com/z-d-g/md-cli/internal/constants"
	"github.com/z-d-g/md-cli/internal/cursor"
	"github.com/z-d-g/md-cli/internal/editor"
	"github.com/z-d-g/md-cli/internal/render"
	"github.com/z-d-g/md-cli/internal/utils"

	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
)

// Model represents the application state
type Model struct {
	Editor         *editor.Editor
	FilePath       string
	Ready          bool
	Width          int
	Height         int
	SavedContent   string
	ShowHelpDialog bool
	HelpAnimOffset float64
	HelpAnimDir    float64
	Notification   string
	Config         *config.Config
	CursorStore    *cursor.PositionStore
	bgIsDark       bool
}

func NewModel(filePath string, cfg *config.Config) Model {
	content := ""
	if data, err := utils.ReadFile(filePath); err == nil {
		content = string(data)
	}

	cursorStore, err := cursor.NewPositionStore()
	if err != nil {
		slog.Warn("cursor position store unavailable", "err", err)
		cursorStore = nil
	}

	edRenderer := render.NewLipglossRenderer(&cfg.EditorStyles)
	ed := editor.NewEditor(content, edRenderer)

	render.SetTableLines(func() []string {
		lines := make([]string, ed.LineCount())
		for i := range lines {
			lines[i] = ed.LineAt(i)
		}
		return lines
	})

	ed.MarkClean()
	ed.SetCursor(0, 0)
	ed.Focus()

	return Model{
		Editor:       ed,
		FilePath:     filePath,
		SavedContent: content,
		Config:       cfg,
		CursorStore:  cursorStore,
	}
}

func NewModelFromContent(content string, cfg *config.Config) Model {
	edRenderer := render.NewLipglossRenderer(&cfg.EditorStyles)
	ed := editor.NewEditor(content, edRenderer)

	render.SetTableLines(func() []string {
		lines := make([]string, ed.LineCount())
		for i := range lines {
			lines[i] = ed.LineAt(i)
		}
		return lines
	})

	ed.MarkClean()
	ed.SetCursor(0, 0)
	ed.Focus()

	return Model{
		Editor:       ed,
		FilePath:     "stdin",
		SavedContent: content,
		Config:       cfg,
		CursorStore:  nil,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Notification == "exit" {
		return m.handleExitConfirm(msg)
	}
	if m.ShowHelpDialog || m.HelpAnimDir != 0.0 {
		if model, cmd, consumed := m.handleHelpDialog(msg); consumed {
			return model, cmd
		}
	}
	return m.handleMessage(msg)
}

func (m Model) handleExitConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "y", "Y", "ctrl+q":
		m.saveCursorPosition()
		m.Editor.MarkClean()
		return m, tea.Quit
	case "n", "N", "esc":
		m.Notification = ""
		return m, nil
	case "ctrl+s":
		if err := m.saveFile(); err != nil {
			m.Notification = "save error"
		} else {
			m.Notification = "saved"
		}
		return m, m.showTemporaryNotification(m.Notification)
	}
	return m, nil
}

func (m Model) handleHelpDialog(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "f1" {
			if m.HelpAnimDir == 0.0 || m.HelpAnimDir == 1.0 {
				m.HelpAnimDir = -1.0
				return m, tickAnim(), true
			}
			return m, nil, true
		}
	case animTickMsg:
		m.updateHelpAnim()
		if m.HelpAnimDir != 0.0 {
			return m, tickAnim(), true
		}
		return m, nil, true
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg)
		return m, nil, true
	}
	if m.HelpAnimDir == 0.0 && m.ShowHelpDialog {
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) handleMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animTickMsg:
		return m.handleAnimTick()
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg)
		return m, nil
	case tea.BackgroundColorMsg:
		return m.handleBackgroundColor(msg)
	case clearNotificationMsg:
		m.clearNotification()
		return m, nil
	case editor.NotificationMsg:
		return m.handleEditorNotification(msg)
	default:
		return m.updateEditor(msg)
	}
}

func (m Model) handleAnimTick() (tea.Model, tea.Cmd) {
	m.updateHelpAnim()
	if m.HelpAnimDir != 0.0 {
		return m, tickAnim()
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q":
		if m.Editor.IsDirty() {
			m.Notification = "exit"
			return m, nil
		}
		m.saveCursorPosition()
		return m, tea.Quit
	case "f1":
		if !m.ShowHelpDialog && m.HelpAnimDir == 0.0 {
			m.ShowHelpDialog = true
			m.HelpAnimDir = 1.0
			return m, tickAnim()
		}
		return m, nil
	case "ctrl+s":
		if err := m.saveFile(); err != nil {
			slog.Error("error saving file", "err", err, "file", m.FilePath)
		} else {
			m.saveCursorPosition()
			m.SavedContent = m.Editor.Value()
			m.Editor.MarkClean()
		}
		m.Notification = ""
		return m, nil
	default:
		return m.updateEditor(msg)
	}
}

func (m Model) handleEditorNotification(msg editor.NotificationMsg) (tea.Model, tea.Cmd) {
	m.Notification = string(msg.MessageType)
	return m, m.showTemporaryNotification(m.Notification)
}

func (m Model) handleBackgroundColor(msg tea.BackgroundColorMsg) (tea.Model, tea.Cmd) {
	m.bgIsDark = msg.IsDark()
	m.Config = config.LoadConfigAdaptive(m.bgIsDark)
	return m, nil
}

func (m *Model) saveFile() error {
	return utils.WriteFile(m.FilePath, []byte(m.Editor.Value()))
}

func (m Model) updateEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	var edCmd tea.Cmd
	newEditor, edCmd := m.Editor.Update(msg)
	if newEditor != nil {
		m.Editor = newEditor.(*editor.Editor)
	}
	if m.Editor.Value() == m.SavedContent {
		m.Editor.MarkClean()
	} else {
		m.Editor.MarkDirty()
	}
	return m, edCmd
}

func (m Model) View() tea.View {
	if !m.Ready {
		v := tea.NewView("\n  Initializing...")
		v.AltScreen = true
		v.MouseMode = tea.MouseModeAllMotion
		return v
	}

	var b strings.Builder
	b.Grow(m.Width * m.Height)

	b.WriteString(m.renderTopBar())
	b.WriteString("\n")

	if m.Editor != nil {
		editorContent := m.Editor.View().Content

		if m.ShowHelpDialog || m.HelpAnimDir != 0.0 {
			editorLines := strings.Split(editorContent, "\n")
			helpDialog := m.renderHelpDialog()
			if len(helpDialog) > 0 {
				helpLines := strings.Split(strings.TrimRight(helpDialog, "\n"), "\n")
				for i := 0; i < len(helpLines) && i < len(editorLines); i++ {
					editorLines[i] = helpLines[i]
				}
			}
			b.WriteString(strings.Join(editorLines, "\n"))
		} else {
			b.WriteString(editorContent)
		}
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func (m *Model) updateHelpAnim() {
	if m.HelpAnimDir == 0 {
		return
	}
	m.HelpAnimOffset += m.HelpAnimDir * constants.AnimationStep
	if m.HelpAnimOffset >= 1.0 {
		m.HelpAnimOffset = 1.0
		m.HelpAnimDir = 0.0
		m.ShowHelpDialog = true
	} else if m.HelpAnimOffset <= 0.0 {
		m.HelpAnimOffset = 0.0
		m.HelpAnimDir = 0.0
		m.ShowHelpDialog = false
	}
}

func (m *Model) applyWindowSize(msg tea.WindowSizeMsg) {
	m.Width = msg.Width
	m.Height = msg.Height
	if !m.Ready {
		m.Ready = true
	}
	editorHeight := max(msg.Height-m.reservedRows(), 0)
	if m.Editor != nil {
		m.Editor.SetSize(msg.Width, editorHeight)
	}
}

func (m *Model) saveCursorPosition() {
	if m.Editor != nil && m.CursorStore != nil {
		currentLine, currentCol := m.Editor.Cursor()
		_ = m.CursorStore.SetPosition(m.FilePath, cursor.FileConfig{
			CursorLine: currentLine,
			CursorCol:  currentCol,
		})
	}
}

func (m Model) reservedRows() int { return 1 }

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}
	if maxWidth == 1 {
		return "…"
	}
	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + "…"
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return "…"
}

func (m Model) renderTopBar() string {
	if m.Width <= 0 {
		return ""
	}

	title := m.Config.TitleStyle.Render("md-cli")
	helpHint := m.Config.HintStyle.Render(" [f1:?]")
	left := title + helpHint

	dirty := m.Editor != nil && m.Editor.IsDirty()
	dotStyle := m.Config.SavedStyle
	if dirty {
		dotStyle = m.Config.UnsavedStyle
	}
	dot := dotStyle.Render("●")

	right := dot
	if m.Notification != "" {
		statusMsg := truncateText(m.getNotificationMessage(), m.Width/4)
		statusStyle := m.Config.InfoStyle
		if m.Notification == "exit" {
			statusStyle = m.Config.UnsavedStyle
		}
		right = statusStyle.Render(statusMsg) + " " + dot
	}

	filename := truncateText(filepath.Base(m.FilePath), m.Width/3)
	center := m.Config.InfoStyle.Render(filename)

	leftW := lipgloss.Width(left)
	centerW := lipgloss.Width(center)
	rightW := lipgloss.Width(right)
	gap := max(m.Width-leftW-rightW, centerW)

	return left + lipgloss.PlaceHorizontal(gap, lipgloss.Center, center) + right
}

func (m Model) getNotificationMessage() string {
	switch m.Notification {
	case string(constants.ExitConfirmation):
		return constants.ExitConfirmation.Message()
	case string(constants.CodeBlockCopied):
		return constants.CodeBlockCopied.Message()
	case string(constants.TextCopied):
		return constants.TextCopied.Message()
	case string(constants.CutToClipboard):
		return constants.CutToClipboard.Message()
	case string(constants.PastedFromClipboard):
		return constants.PastedFromClipboard.Message()
	default:
		return m.Notification
	}
}

func easeOutQuart(x float64) float64 {
	return 1.0 - (1.0-x)*(1.0-x)*(1.0-x)*(1.0-x)
}

// renderHelpDialog renders the help dialog.
// Groups are arranged in two rows of columns, wrapped in a lipgloss border (left, bottom, right).
// Animation: pulls down from top.
func (m Model) renderHelpDialog() string {
	if m.Editor == nil || m.Width < 30 {
		return ""
	}

	easedOffset := easeOutQuart(m.HelpAnimOffset)
	if easedOffset <= 0 {
		return ""
	}

	key := m.Config.HelpKeyStyle
	desc := m.Config.HelpDescStyle
	section := m.Config.HelpSectionStyle
	info := m.Config.InfoStyle

	// ── Group data ──
	type kv = [2]string
	type group struct {
		title string
		items []kv
	}

	groups := []group{
		{"File", []kv{{"Ctrl+S", "Save"}, {"Ctrl+Q", "Quit"}, {"", "STATUS"}}},
		{"Cursor", []kv{{"↑↓←→", "Move"}, {"Home/End", "Line"}, {"Ctrl+←→", "Word"}, {"PgUp/PgDn", "Page"}, {"Ctrl+Home/End", "Doc"}}},
		{"Select", []kv{{"Shift+↑↓←→", "Extend"}, {"Ctrl+A", "All"}, {"Ctrl+D", "Word"}, {"Ctrl+L", "Line"}}},
		{"Edit", []kv{{"Ctrl+C", "Copy"}, {"Ctrl+X", "Cut"}, {"Ctrl+V", "Paste"}, {"Ctrl+Z/Y", "Undo/Redo"}, {"Bksp/Del", "Delete"}, {"Alt+Bksp/Del", "Word del"}}},
	}

	// ── Build content lines ──
	innerPad := 2
	borderW := 2 // lipgloss RoundedBorder: 1 left + 1 right
	contentWidth := m.Width - borderW - innerPad*2

	// Find max key width for alignment
	maxKeyW := 0
	for _, g := range groups {
		for _, item := range g.items {
			maxKeyW = max(maxKeyW, lipgloss.Width(key.Render(item[0])))
		}
	}

	pair := func(k, d string) string {
		if k == "" && d == "STATUS" {
			return m.Config.SavedStyle.Render("● saved") + "  " + m.Config.UnsavedStyle.Render("● unsaved")
		}
		kRendered := key.Render(k)
		pad := maxKeyW - lipgloss.Width(kRendered)
		return kRendered + strings.Repeat(" ", max(1, pad+1)) + desc.Render(strings.ToLower(d))
	}

	// ── Layout: columns if wide enough, vertical stack if narrow ──
	var contentLines []string

	if contentWidth >= 83 {
		// Column layout
		colW := (contentWidth - (len(groups) - 1)) / len(groups)

		// Headers
		var headerLine strings.Builder
		for i, g := range groups {
			if i > 0 {
				headerLine.WriteString(" ")
			}
			tag := " " + g.title + " "
			h := section.Render("─" + tag + "─")
			headerLine.WriteString(h)
			headerPad := colW - lipgloss.Width(h)
			if headerPad > 0 {
				headerLine.WriteString(strings.Repeat(" ", headerPad))
			}
		}
		contentLines = append(contentLines, strings.Repeat(" ", innerPad)+headerLine.String())

		// Items interleaved
		maxItems := 0
		for _, g := range groups {
			maxItems = max(maxItems, len(g.items))
		}
		for row := 0; row < maxItems; row++ {
			var line strings.Builder
			for ci, g := range groups {
				if ci > 0 {
					line.WriteString(" ")
				}
				if row < len(g.items) {
					cell := pair(g.items[row][0], g.items[row][1])
					line.WriteString(cell)
					pad := colW - lipgloss.Width(cell)
					if pad > 0 {
						line.WriteString(strings.Repeat(" ", pad))
					}
				} else {
					line.WriteString(strings.Repeat(" ", colW))
				}
			}
			contentLines = append(contentLines, strings.Repeat(" ", innerPad)+line.String())
		}
	} else {
		// Vertical stack layout
		for _, g := range groups {
			tag := " " + g.title + " "
			contentLines = append(contentLines, strings.Repeat(" ", innerPad)+section.Render("─"+tag+"─"))
			for _, item := range g.items {
				contentLines = append(contentLines, strings.Repeat(" ", innerPad)+pair(item[0], item[1]))
			}
			contentLines = append(contentLines, strings.Repeat(" ", innerPad))
		}
	}

	// Footer
	footerText := " esc or f1 to close "
	dashW := max(contentWidth-lipgloss.Width(footerText), 0)
	contentLines = append(contentLines, strings.Repeat(" ", innerPad)+info.Render(
		strings.Repeat("─", dashW/2)+footerText+strings.Repeat("─", dashW-dashW/2),
	))

	// ── Apply borders: left, bottom, right (no top) ──
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(m.Config.ModalBorderStyle.GetBorderBottomForeground()).
		Render(strings.Join(contentLines, "\n"))

	// ── Animate: pull down from top ──
	dialogLines := strings.Split(dialog, "\n")
	visibleCount := int(easedOffset * float64(len(dialogLines)))
	if visibleCount == 0 {
		return ""
	}

	var result strings.Builder
	for _, line := range dialogLines[:visibleCount] {
		result.WriteString(line)
		result.WriteByte('\n')
	}
	return result.String()
}

func (m *Model) showTemporaryNotification(_ string) tea.Cmd {
	return tea.Tick(constants.NotificationDuration, func(t time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}

func (m *Model) clearNotification() {
	m.Notification = ""
}

type animTickMsg time.Time

func tickAnim() tea.Cmd {
	return tea.Tick(constants.AnimationInterval, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

type clearNotificationMsg struct{}
