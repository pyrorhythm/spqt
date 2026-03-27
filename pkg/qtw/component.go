package qtw

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

// Widgeter is implemented by anything that wraps a QWidget — components,
// builders, etc. Layout Items() accepts Widgeters directly.
type Widgeter interface {
	Widget() *qt.QWidget
}

// Disposable is implemented by components that need cleanup.
type Disposable interface {
	Dispose()
}

// Component is a base struct that concrete UI components embed.
// It provides a root widget, child tracking, and reactive Watch support.
type Component struct {
	root     *qt.QWidget
	children []Widgeter
	unsubs   []func()
}

// Root sets the component's root widget.
func (c *Component) Root(w *qt.QWidget) *Component {
	c.root = w
	return c
}

// Widget returns the component's root QWidget.
func (c *Component) Widget() *qt.QWidget {
	return c.root
}

// Child registers a sub-component for future lifecycle management.
// Returns the child for inline use.
func (c *Component) Child(child Widgeter) Widgeter {
	c.children = append(c.children, child)
	return child
}

// WatchUnsub registers an unsubscribe function for cleanup on Dispose.
func (c *Component) WatchUnsub(unsub func()) {
	c.unsubs = append(c.unsubs, unsub)
}

// Dispose unsubscribes all listeners, disposes children, then destroys the root widget.
func (c *Component) Dispose() {
	for _, fn := range c.unsubs {
		fn()
	}
	c.unsubs = nil
	for _, child := range c.children {
		if d, ok := child.(Disposable); ok {
			d.Dispose()
		}
	}
	if c.root != nil {
		c.root.SetParent(nil)
		c.root.Delete()
		c.root = nil
	}
}

// Watch subscribes to an Observable, calling fn immediately with the current
// value and on every subsequent change.
// Returns an unsubscribe function.
func Watch[T any](obs reactive.Observable[T], fn func(T)) func() {
	fn(obs.Get())
	return obs.OnChange(func(v T) {
		fn(v)
	})
}
