package reactive

type List[T any] struct {
	items     []T
	listeners []func()
}

func NewList[T any]() *List[T] {
	return &List[T]{}
}

func (l *List[T]) Items() []T {
	return l.items
}

func (l *List[T]) Len() int {
	return len(l.items)
}

func (l *List[T]) At(i int) T {
	return l.items[i]
}

func (l *List[T]) Append(v T) {
	l.items = append(l.items, v)
	l.notify()
}

func (l *List[T]) Remove(i int) {
	l.items = append(l.items[:i], l.items[i+1:]...)
	l.notify()
}

func (l *List[T]) Set(i int, v T) {
	l.items[i] = v
	l.notify()
}

func (l *List[T]) Clear() {
	l.items = l.items[:0]
	l.notify()
}

func (l *List[T]) SetItems(items []T) {
	l.items = items
	l.notify()
}

func (l *List[T]) OnChange(fn func()) {
	l.listeners = append(l.listeners, fn)
}

func (l *List[T]) notify() {
	for _, fn := range l.listeners {
		fn()
	}
}
