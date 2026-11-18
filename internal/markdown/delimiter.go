package markdown

import "strings"

// CodeSpanClose returns the index of the closing backtick run for a code span
// opened at start. A run closes the span only when its length is exactly runLen
// (CommonMark: the closing run must match the opening run in length).
func CodeSpanClose(line string, start, runLen int) int {
	if runLen <= 0 {
		return -1
	}
	for i := start; i < len(line); {
		if line[i] != '`' {
			i++
			continue
		}
		j := i
		for j < len(line) && line[j] == '`' {
			j++
		}
		if j-i == runLen {
			return i
		}
		i = j
	}
	return -1
}

// IsEscaped reports whether the byte at index i is preceded by an odd number of
// backslashes.
func IsEscaped(line string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// SplitRun returns the length of the run of line[i] starting at i.
func SplitRun(line string, i int, run string) int {
	if i < 0 || i >= len(line) || strings.IndexByte(run, line[i]) < 0 {
		return 0
	}
	j := i
	for j < len(line) && line[j] == line[i] {
		j++
	}
	return j - i
}

// FindClosingDelimiter finds the closing delimiter in line starting from position
// start, where start is the index just past the opening delimiter.
// Returns -1 when there is no valid closer.
func FindClosingDelimiter(line string, start int, delimiter string) int {
	if delimiter == "" || start < 0 || start > len(line) {
		return -1
	}
	switch {
	case delimiter == "`":
		return CodeSpanClose(line, start, 1)
	case delimiter == "*" || delimiter == "_":
		return emphasisClose(line, start, delimiter[0])
	case delimiter == "**" || delimiter == "__":
		return emphasisClose(line, start, delimiter[0])
	case delimiter == "***" || delimiter == "___":
		return emphasisClose(line, start, delimiter[0])
	}

	delimLen := len(delimiter)
	for i := start; i+delimLen <= len(line); i++ {
		if line[i] != delimiter[0] || IsEscaped(line, i) {
			continue
		}
		if line[i:i+delimLen] == delimiter {
			return i
		}
	}
	return -1
}

// emphasisClose finds the run of ch that closes an emphasis opener ending at
// from. Only a run of the same character can close, so `*a*` and `**b**` match
// their own lengths.
func emphasisClose(line string, from int, ch byte) int {
	for i := from; i < len(line); {
		if line[i] != ch || IsEscaped(line, i) {
			i++
			continue
		}
		j := i
		for j < len(line) && line[j] == ch {
			j++
		}
		if isClosingRun(line, i, j-i, ch) {
			return i
		}
		i = j
	}
	return -1
}

// isClosingRun reports whether the run at start may close emphasis. A run must
// not be followed by the same character (that would be a longer run) and must
// not be preceded by a backslash.
func isClosingRun(line string, start, length int, ch byte) bool {
	if IsEscaped(line, start) {
		return false
	}
	if start+length < len(line) && line[start+length] == ch {
		return false
	}
	return true
}

func isPunctByte(c byte) bool {
	return strings.IndexByte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", c) >= 0
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
