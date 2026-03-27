package reactive

// Derive creates a Prop[T] whose value is computed from a source Observable.
// The returned Prop fires only when the derived value actually changes.
// The second return value is an unsubscribe function; call it to stop the derivation.
func Derive[S any, T comparable](source Observable[S], fn func(S) T) (*Prop[T], func()) {
	p := NewProp(fn(source.Get()))
	unsub := source.OnChange(func(s S) {
		p.Set(fn(s))
	})
	return p, unsub
}

// Always returns a Prop that holds a constant value and never changes.
func Always[T comparable](v T) *Prop[T] {
	return NewProp(v)
}
