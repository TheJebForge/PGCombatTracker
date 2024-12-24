package utils

// AbsInt "if it's simple to write, it shouldn't be in stdlib" my ass!
func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
