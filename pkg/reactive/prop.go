package reactive

type Prop[T comparable] struct {
	value          T
	listeners      []func(T)
	exactListeners map[T][]func()
}

func NewProp[T comparable](initial T) *Prop[T] {
	return &Prop[T]{value: initial, exactListeners: make(map[T][]func())}
}

func (p *Prop[T]) Get() T {
	return p.value
}

func (p *Prop[T]) Set(v T) {
	if p.value == v {
		return
	}
	p.value = v

	for _, efn := range p.exactListeners[v] {
		efn()
	}
	for _, fn := range p.listeners {
		fn(v)
	}
}

func (p *Prop[T]) OnChange(fn func(T)) {
	p.listeners = append(p.listeners, fn)
}

func (p *Prop[T]) OnExact(wanted T, fn func()) {
	p.exactListeners[wanted] = append(p.exactListeners[wanted], fn)
}
