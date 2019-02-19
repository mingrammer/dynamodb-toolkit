package calc

// Min returns a smaller value
func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
