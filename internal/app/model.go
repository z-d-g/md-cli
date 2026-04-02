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
	"github.com/mattn/go-runewidth"
)

type modelState int

const (
	stateNormal modelState = iota
	stateExitConfirm
	stateHelpDialog
)

type Model struct {
	Editor         *editor.Editor
	FilePath       string
	Ready          bool
	Width          int
	Height         int
	SavedContent   string
	state          modelState
	HelpAnimOffset float64
	HelpAnimDir    float64
	Notification   string
	Config         *config.Config
	CursorStore    *cursor.PositionStore

	dialogLines []string
	dialogWidth int
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

	edRenderer := render.NewLipglossRenderer(&cfg.EditorStyles, 0)
	ed := editor.NewEditor(content, edRenderer)

	edRenderer.SetDocument(func() []string {
		lines := make([]string, ed.LineCount())
		for i := range lines {
			lines[i] = ed.LineAt(i)
		}
		return lines
	})

	ed.MarkClean()
	if cursorStore != nil {
		if pos, ok := cursorStore.GetPosition(filePath); ok {
			ed.SetCursor(pos.CursorLine, pos.CursorCol)
		}
	}
	ed.Focus()

	return Model{
		Editor:       ed,
		FilePath:     filePath,
		SavedContent: content,
		Config:       cfg,
		CursorStore:  cursorStore,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateExitConfirm:
		return m.handleExitConfirm(msg)
	case stateHelpDialog:
		if model, cmd, consumed := m.handleHelpDialog(msg); consumed {
			return model, cmd
		}
	}
	return m.handleMessage(msg)
}

func (m Model) handleExitConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y", "ctrl+q":
			m.saveCursorPosition()
			m.Editor.MarkClean()
			return m, tea.Quit
		case "n", "N", "esc":
			m.Notification = ""
			m.state = stateNormal
			return m, nil
		case "ctrl+s":
			if err := m.saveFile(); err != nil {
				slog.Error("error saving file", "err", err, "file", m.FilePath)
				m.Notification = "save error"
			} else {
				m.saveCursorPosition()
				m.SavedContent = m.Editor.Value()
				m.Editor.MarkClean()
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg)
		return m, nil
	}
	return m, nil
}

func (m Model) handleHelpDialog(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "esc" || msg.String() == "f1" {
			m.HelpAnimDir = -1.0
		}
	case animTickMsg:
		wasOpening := m.HelpAnimDir > 0
		m.updateHelpAnim()
		if m.HelpAnimDir != 0.0 {
			return m, tickAnim(), true
		}
		if wasOpening {
			return m, nil, true
		}
		m.state = stateNormal
		return m, nil, true
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg)
	}

	if m.HelpAnimDir != 0.0 {
		return m, tickAnim(), true
	}
	return m, nil, true
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
			m.Notification = string(constants.ExitConfirmation)
			m.state = stateExitConfirm
			return m, nil
		}
		m.saveCursorPosition()
		return m, tea.Quit
	case "f1":
		if m.state != stateHelpDialog && m.HelpAnimDir == 0.0 {
			m.state = stateHelpDialog
			m.HelpAnimOffset = 0.0
			m.HelpAnimDir = 1.0
			return m, tickAnim()
		}
		if m.state == stateHelpDialog {
			m.HelpAnimDir = -1.0
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
		return m, nil
	default:
		return m.updateEditor(msg)
	}
}

func (m Model) handleEditorNotification(msg editor.NotificationMsg) (tea.Model, tea.Cmd) {
	m.Notification = string(msg.MessageType)
	return m, m.showTemporaryNotification(m.Notification)
}

func (m Model) handleBackgroundColor(_ tea.BackgroundColorMsg) (tea.Model, tea.Cmd) {
	m.Config = config.LoadConfig()
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
		if m.state == stateHelpDialog || m.HelpAnimDir != 0.0 {
			if overlay := m.renderHelpOverlay(editorContent); overlay != "" {
				b.WriteString(overlay)
			} else {
				b.WriteString(editorContent)
			}
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
	} else if m.HelpAnimOffset <= 0.0 {
		m.HelpAnimOffset = 0.0
		m.HelpAnimDir = 0.0
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
	return truncWithEllipsis(text, maxWidth)
}

const statusMinWidth = 28

func (m Model) statusMessage() string {
	msg := m.getNotificationMessage()
	budget := max(m.Width/4, statusMinWidth)
	return truncateText(msg, budget)
}

func truncWithEllipsis(text string, maxWidth int) string {
	suffix := "…"
	budget := maxWidth - 1
	if budget <= 0 {
		return suffix
	}
	var totalW int
	for i, r := range text {
		if totalW+runewidth.RuneWidth(r) > budget {
			return text[:i] + suffix
		}
		totalW += runewidth.RuneWidth(r)
	}
	return text
}

func (m Model) renderTopBar() string {
	if m.Width <= 0 {
		return ""
	}

	title := m.Config.TitleStyle.Render("md-cli")
	helpHint := m.Config.HintStyle.Render(" [f1:?]")
	leftLabel := title + helpHint

	dirty := m.Editor != nil && m.Editor.IsDirty()
	dotStyle := m.Config.SavedStyle
	if dirty {
		dotStyle = m.Config.UnsavedStyle
	}
	dot := dotStyle.Render("●")

	right := dot
	if m.Notification != "" {
		statusMsg := m.statusMessage()
		statusStyle := m.Config.InfoStyle
		if m.state == stateExitConfirm {
			statusStyle = m.Config.UnsavedStyle
		}
		right = statusStyle.Render(statusMsg) + " " + dot
	}

	filename := truncateText(filepath.Base(m.FilePath), m.Width/2)
	fileNameStyled := m.Config.InfoStyle.Render(filename)

	leftW := lipgloss.Width(leftLabel)
	rightW := lipgloss.Width(right)
	remaining := m.Width - leftW - rightW - 3

	if remaining > 20 {
		fillW := max(remaining-lipgloss.Width(fileNameStyled)-1, 1)
		fill := m.Config.HintStyle.Render(strings.Repeat("─", fillW))
		return leftLabel + " " + fileNameStyled + " " + fill + " " + right
	}

	fillW := max(remaining-lipgloss.Width(fileNameStyled), 0)
	if fillW > 0 {
		fill := m.Config.HintStyle.Render(strings.Repeat("─", fillW))
		return leftLabel + " " + fill + " " + right
	}

	gap := max(m.Width-leftW-rightW, 1)
	return leftLabel + lipgloss.PlaceHorizontal(gap, lipgloss.Right, "") + right
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

func (m *Model) helpDialogLines() []string {
	if len(m.dialogLines) == 0 || m.Width != m.dialogWidth {
		dialog := m.renderHelpDialogContent(m.Config.HelpKeyStyle, m.Config.HelpDescStyle, m.Config.HelpSectionStyle, m.Config.InfoStyle)
		m.dialogWidth = m.Width
		if dialog == "" {
			m.dialogLines = nil
		} else {
			m.dialogLines = strings.Split(strings.TrimRight(dialog, "\n"), "\n")
		}
	}
	return m.dialogLines
}

func (m Model) renderHelpOverlay(editorContent string) string {
	if m.Editor == nil || m.Width < 30 || m.Height <= 0 {
		return editorContent
	}

	dialogLines := m.helpDialogLines()
	if len(dialogLines) == 0 {
		return editorContent
	}

	viewport := m.Height - m.reservedRows()

	topOffset := int((1.0 - easeOutQuart(m.HelpAnimOffset)) * float64(len(dialogLines)))
	if topOffset >= len(dialogLines) {
		return editorContent
	}

	lines := make([]string, 0, viewport)
	for i := topOffset; i < len(dialogLines) && len(lines) < viewport; i++ {
		lines = append(lines, dialogLines[i])
	}

	startEditor := len(lines)
	editorLines := strings.Split(editorContent, "\n")
	for i := startEditor; i < len(editorLines) && len(lines) < viewport; i++ {
		lines = append(lines, editorLines[i])
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderHelpDialogContent(key, desc, section, info lipgloss.Style) string {
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

	padL := 2
	padR := 1
	borderW := 2
	contentWidth := m.Width - borderW - padL - padR

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

	var numCols int
	switch {
	case contentWidth >= 83:
		numCols = 4
	case contentWidth >= 55:
		numCols = 2
	default:
		numCols = 1
	}

	var contentLines []string

	if numCols == 1 {
		for _, g := range groups {
			tag := " " + g.title + " "
			contentLines = append(contentLines,
				strings.Repeat(" ", padL)+section.Render("─"+tag+"─")+strings.Repeat(" ", padR))
			for _, item := range g.items {
				contentLines = append(contentLines,
					strings.Repeat(" ", padL)+pair(item[0], item[1])+strings.Repeat(" ", padR))
			}
			contentLines = append(contentLines,
				strings.Repeat(" ", padL)+strings.Repeat(" ", padR))
		}
	} else {
		rowGroups := chunkGroups(len(groups), numCols)
		for _, indices := range rowGroups {
			colW := (contentWidth - (len(indices) - 1)) / len(indices)

			var headerBuf strings.Builder
			for i, gi := range indices {
				if i > 0 {
					headerBuf.WriteString(" ")
				}
				tag := " " + groups[gi].title + " "
				h := section.Render("─" + tag + "─")
				headerBuf.WriteString(h)
				if p := colW - lipgloss.Width(h); p > 0 {
					headerBuf.WriteString(strings.Repeat(" ", p))
				}
			}
			contentLines = append(contentLines,
				strings.Repeat(" ", padL)+headerBuf.String()+strings.Repeat(" ", padR))

			maxItems := 0
			for _, gi := range indices {
				maxItems = max(maxItems, len(groups[gi].items))
			}
			for row := 0; row < maxItems; row++ {
				var lineBuf strings.Builder
				for i, gi := range indices {
					if i > 0 {
						lineBuf.WriteString(" ")
					}
					if row < len(groups[gi].items) {
						cell := pair(groups[gi].items[row][0], groups[gi].items[row][1])
						lineBuf.WriteString(cell)
						if p := colW - lipgloss.Width(cell); p > 0 {
							lineBuf.WriteString(strings.Repeat(" ", p))
						}
					} else {
						lineBuf.WriteString(strings.Repeat(" ", colW))
					}
				}
				contentLines = append(contentLines,
					strings.Repeat(" ", padL)+lineBuf.String()+strings.Repeat(" ", padR))
			}
		}
	}

	footerText := " esc or f1 to close "
	dashW := max(contentWidth-lipgloss.Width(footerText), 0)
	footerContent := info.Render(
		strings.Repeat("─", dashW/2) + footerText + strings.Repeat("─", dashW-dashW/2),
	)
	contentLines = append(contentLines,
		strings.Repeat(" ", padL)+footerContent+strings.Repeat(" ", padR))

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(m.Config.ModalBorderStyle.GetBorderBottomForeground()).
		Render(strings.Join(contentLines, "\n"))

	return dialog
}

func chunkGroups(n, cols int) [][]int {
	rows := (n + cols - 1) / cols
	result := make([][]int, 0, rows)
	for r := range rows {
		start := r * cols
		end := min(start+cols, n)
		chunk := make([]int, end-start)
		for i := start; i < end; i++ {
			chunk[i-start] = i
		}
		result = append(result, chunk)
	}
	return result
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
