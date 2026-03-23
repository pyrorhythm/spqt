package qtw

import qt "github.com/mappu/miqt/qt6"

// Pages is a named wrapper around QStackedWidget.
type Pages struct {
	stack *qt.QStackedWidget
	index map[string]int
}

// NewPages creates a named stacked widget.
func NewPages() *Pages {
	return &Pages{
		stack: qt.NewQStackedWidget2(),
		index: make(map[string]int),
	}
}

// Page adds a named page. Returns self for chaining.
func (p *Pages) Page(name string, widget *qt.QWidget) *Pages {
	idx := p.stack.AddWidget(widget)
	p.index[name] = idx
	return p
}

// Show switches to the named page.
func (p *Pages) Show(name string) {
	if idx, ok := p.index[name]; ok {
		p.stack.SetCurrentIndex(idx)
	}
}

// CurrentName returns the name of the current page, or "" if unknown.
func (p *Pages) CurrentName() string {
	cur := p.stack.CurrentIndex()
	for name, idx := range p.index {
		if idx == cur {
			return name
		}
	}
	return ""
}

// Widget returns the underlying QWidget for layout embedding.
func (p *Pages) Widget() *qt.QWidget {
	return p.stack.QWidget
}

// Raw returns the underlying QStackedWidget.
func (p *Pages) Raw() *qt.QStackedWidget {
	return p.stack
}
