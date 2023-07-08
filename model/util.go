package m

func isNewLine(r rune) bool {
	return r == '\n'
}

func min(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}
