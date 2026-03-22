package reactive

type Prop[T comparable] struct {
	value     T
	listeners []func(T)
}

func NewProp[T comparable](initial T) *Prop[T] {
	return &Prop[T]{value: initial}
}

func (p *Prop[T]) Get() T {
	return p.value
}

func (p *Prop[T]) Set(v T) {
	if p.value == v {
		return
	}
	p.value = v
	for _, fn := range p.listeners {
		fn(v)
	}
}

func (p *Prop[T]) OnChange(fn func(T)) {
	p.listeners = append(p.listeners, fn)
}
