package reactive

type Signal[T any] struct {
	listeners []func(T)
}

func NewSignal[T any]() *Signal[T] {
	return &Signal[T]{}
}

func (s *Signal[T]) Emit(v T) {
	for _, fn := range s.listeners {
		fn(v)
	}
}

func (s *Signal[T]) Subscribe(fn func(T)) {
	s.listeners = append(s.listeners, fn)
}

type VoidSignal struct {
	listeners []func()
}

func NewVoidSignal() *VoidSignal {
	return &VoidSignal{}
}

func (s *VoidSignal) Emit() {
	for _, fn := range s.listeners {
		fn()
	}
}

func (s *VoidSignal) Subscribe(fn func()) {
	s.listeners = append(s.listeners, fn)
}
