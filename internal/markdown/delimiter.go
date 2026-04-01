package markdown

// FindClosingDelimiter finds the closing delimiter in line starting from position start.
// Returns -1 if no closing delimiter is found.
func FindClosingDelimiter(line string, start int, delimiter string) int {
	if delimiter == "" {
		return -1
	}

	// For backticks, handle multiple backticks case
	if delimiter == "`" {
		backtickCount := len(delimiter)
		for i := start; i < len(line); i++ {
			if line[i] == '`' {
				count := 1
				for j := i + 1; j < len(line) && line[j] == '`'; j++ {
					count++
				}
				if count >= backtickCount {
					return i
				}
			}
		}
		return -1
	}

	// For other delimiters
	delimLen := len(delimiter)
	for i := start; i < len(line); {
		if i+delimLen <= len(line) && line[i:i+delimLen] == delimiter {
			// Ensure it's not escaped
			if i > 0 && line[i-1] == '\\' {
				i++
				continue
			}

			// Check delimiter run rules for * and _
			if delimiter == "*" || delimiter == "_" {
				if delimLen == 1 {
					if i+2 <= len(line) && line[i:i+2] == "**" {
						i += 2
						continue
					}
					if i+3 <= len(line) && line[i:i+3] == "***" {
						i += 3
						continue
					}
				}
			}

			if (delimiter == "**" || delimiter == "__") && delimLen == 2 {
				if i+3 <= len(line) && (line[i:i+3] == "***" || line[i:i+3] == "___") {
					i += 3
					continue
				}
			}

			return i
		}
		i++
	}

	return -1
}
