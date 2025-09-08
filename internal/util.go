package internal

// ordered is a constraint that permits any ordered type.
type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// min returns the smaller of a or b.
func min[T ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of a or b.
func max[T ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
