package reactive

// Observable is the read-only contract shared by Prop, CmpProp, and Derived values.
type Observable[T any] interface {
	Get() T
	OnChange(func(T)) func()
}
