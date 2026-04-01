package constants

type NotificationType string

const (
	CodeBlockCopied     NotificationType = "code_copied"
	TextCopied          NotificationType = "text_copied"
	CutToClipboard      NotificationType = "cut_to_clipboard"
	PastedFromClipboard NotificationType = "pasted_from_clipboard"
	ExitConfirmation    NotificationType = "exit"
)

func (nt NotificationType) Message() string {
	switch nt {
	case CodeBlockCopied:
		return "Copied code block"
	case TextCopied:
		return "Copied text"
	case CutToClipboard:
		return "Cut to clipboard"
	case PastedFromClipboard:
		return "Pasted"
	case ExitConfirmation:
		return "Unsaved changes. Quit? (y/n)"
	default:
		return string(nt)
	}
}
