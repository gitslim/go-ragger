package util

// Of returns a pointer to the given value
func Of[T any](t T) *T {
	return &t
}
