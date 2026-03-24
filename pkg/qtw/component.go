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

// Component is a base struct that concrete UI components embed.
// It provides a root widget, child tracking, and reactive Watch support.
type Component struct {
	root     *qt.QWidget
	children []Widgeter
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

// Watch subscribes to an Observable, calling fn immediately with the current
// value and on every subsequent change. Free function because Go disallows
// generic methods on non-generic types.
func Watch[T any](obs reactive.Observable[T], fn func(T)) {
	fn(obs.Get())
	obs.OnChange(fn)
}
