package reactive

// Derive creates a Prop[T] whose value is computed from a source Observable.
// The returned Prop fires only when the derived value actually changes.
func Derive[S any, T comparable](source Observable[S], fn func(S) T) *Prop[T] {
	p := NewProp(fn(source.Get()))
	source.OnChange(func(s S) {
		p.Set(fn(s))
	})
	return p
}

// Always returns a Prop that holds a constant value and never changes.
func Always[T comparable](v T) *Prop[T] {
	return NewProp(v)
}
