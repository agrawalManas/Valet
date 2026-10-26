package CustomUtils

// Generic helper that safely dereferences a pointer
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
