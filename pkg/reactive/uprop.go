package reactive

type Comparator[T any, K comparable] interface {
	Compare(a, b T) int
	Key(a T) K
}

type CmpProp[T any, K comparable] struct {
	value          T
	cmp            Comparator[T, K]
	listeners      []func(T)
	exactListeners map[K][]func()
}

func NewUProp[T any, K comparable](initial T, cmp Comparator[T, K]) *CmpProp[T, K] {
	return &CmpProp[T, K]{value: initial, cmp: cmp, exactListeners: make(map[K][]func())}
}

func (p *CmpProp[T, _]) Get() T {
	return p.value
}

func (p *CmpProp[T, _]) Set(v T) {
	if p.cmp.Compare(v, p.value) == 0 { // 0 - eq
		return
	}
	p.value = v

	for _, efn := range p.exactListeners[p.cmp.Key(v)] {
		efn()
	}
	for _, fn := range p.listeners {
		fn(v)
	}
}

func (p *CmpProp[T, _]) OnChange(fn func(T)) {
	p.listeners = append(p.listeners, fn)
}

func (p *CmpProp[T, _]) OnExact(wanted T, fn func()) {
	p.exactListeners[p.cmp.Key(wanted)] = append(p.exactListeners[p.cmp.Key(wanted)], fn)
}
