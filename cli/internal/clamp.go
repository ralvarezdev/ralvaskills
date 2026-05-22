package internal

// ClampInt returns n if it is between lo and hi, or the nearest bound if it is outside that range.
func ClampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}