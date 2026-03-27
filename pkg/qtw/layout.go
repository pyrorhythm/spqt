package qtw

import (
	"reflect"

	qt "github.com/mappu/miqt/qt6"
)

type stretchItem struct{ factor int }
type spacingItem struct{ px int }

// Stretch adds a stretchable space in a box layout.
func Stretch() stretchItem { return stretchItem{0} }

// StretchN adds a stretchable space with the given factor.
func StretchN(n int) stretchItem { return stretchItem{n} }

// Space adds a fixed-pixel spacer.
func Space(px int) spacingItem { return spacingItem{px} }

// BoxBuilder configures a QHBoxLayout or QVBoxLayout via chaining.
type BoxBuilder struct {
	box       *qt.QBoxLayout
	stretches [][2]int // deferred Stretch(index, factor) calls
}

// VBox creates a vertical box layout builder.
func VBox() *BoxBuilder {
	return &BoxBuilder{box: qt.NewQVBoxLayout2().QBoxLayout}
}

// HBox creates a horizontal box layout builder.
func HBox() *BoxBuilder {
	return &BoxBuilder{box: qt.NewQHBoxLayout2().QBoxLayout}
}

func (b *BoxBuilder) Spacing(px int) *BoxBuilder {
	b.box.SetSpacing(px)
	return b
}

func (b *BoxBuilder) Margins(left, top, right, bottom int) *BoxBuilder {
	b.box.SetContentsMargins(left, top, right, bottom)
	return b
}

func (b *BoxBuilder) NoMargins() *BoxBuilder {
	return b.Margins(0, 0, 0, 0)
}

// Items adds children and returns the built *qt.QLayout.
// Accepted types: *qt.QWidget, any miqt widget (has .QWidget field),
// *qt.QLayout, *BoxBuilder, *GridBuilder, stretchItem, spacingItem.
func (b *BoxBuilder) Items(items ...any) *qt.QLayout {
	for _, item := range items {
		addBoxItem(b.box, item)
	}
	for _, s := range b.stretches {
		b.box.SetStretch(s[0], s[1])
	}
	return b.box.QLayout
}

// Q returns the layout without adding items.
func (b *BoxBuilder) Q() *qt.QLayout {
	return b.box.QLayout
}

// Box returns the box without adding items.
func (b *BoxBuilder) Box() *qt.QBoxLayout {
	return b.box
}

// Stretch sets stretch factor for item at index (applied after Items()).
func (b *BoxBuilder) Stretch(index, factor int) *BoxBuilder {
	b.stretches = append(b.stretches, [2]int{index, factor})
	return b
}

// GridBuilder configures a QGridLayout via chaining.
type GridBuilder struct {
	grid *qt.QGridLayout
	cols int
	row  int
	col  int
}

// Grid creates a grid layout builder.
func Grid() *GridBuilder {
	return &GridBuilder{grid: qt.NewQGridLayout2(), cols: 1}
}

func (g *GridBuilder) Cols(n int) *GridBuilder {
	g.cols = n
	return g
}

func (g *GridBuilder) Spacing(px int) *GridBuilder {
	g.grid.SetSpacing(px)
	return g
}

func (g *GridBuilder) Margins(left, top, right, bottom int) *GridBuilder {
	g.grid.SetContentsMargins(left, top, right, bottom)
	return g
}

// Items fills the grid left-to-right, top-to-bottom, wrapping at Cols().
func (g *GridBuilder) Items(items ...any) *qt.QLayout {
	for _, item := range items {
		w := mustWidget(item)
		if w != nil {
			g.grid.AddWidget2(w, g.row, g.col)
		}
		g.col++
		if g.col >= g.cols {
			g.col = 0
			g.row++
		}
	}
	return g.grid.QLayout
}

// Put places a widget at a specific row, col.
func (g *GridBuilder) Put(item any, row, col int) *GridBuilder {
	if w := mustWidget(item); w != nil {
		g.grid.AddWidget2(w, row, col)
	}
	return g
}

// PutSpan places a widget at row, col spanning rowSpan x colSpan cells.
func (g *GridBuilder) PutSpan(item any, row, col, rowSpan, colSpan int) *GridBuilder {
	if w := mustWidget(item); w != nil {
		g.grid.AddWidget3(w, row, col, rowSpan, colSpan)
	}
	return g
}

func (g *GridBuilder) Q() *qt.QLayout {
	return g.grid.QLayout
}

func addBoxItem(box *qt.QBoxLayout, v any) {
	switch w := v.(type) {
	case stretchItem:
		if w.factor > 0 {
			box.AddStretchWithStretch(w.factor)
		} else {
			box.AddStretch()
		}
	case spacingItem:
		box.AddSpacing(w.px)
	case *qt.QWidget:
		box.AddWidget(w)
	case *qt.QLayout:
		box.AddLayout(w)
	case Widgeter:
		box.AddWidget(w.Widget())
	case *BoxBuilder:
		box.AddLayout(w.box.QLayout)
	case *GridBuilder:
		box.AddLayout(w.grid.QLayout)
	default:
		if w := extract[qt.QWidget](v, "QWidget"); w != nil {
			box.AddWidget(w)
			return
		}
		if l := extract[qt.QLayout](v, "QLayout"); l != nil {
			box.AddLayout(l)
			return
		}
	}
}

func mustWidget(v any) *qt.QWidget {
	switch w := v.(type) {
	case *qt.QWidget:
		return w
	case Widgeter:
		return w.Widget()
	default:
		return extract[qt.QWidget](v, "QWidget")
	}
}

func extract[T any](v any, field string) *T {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.IsNil() {
		return nil
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	f := rv.FieldByName(field)
	if !f.IsValid() {
		return nil
	}
	if l, ok := f.Interface().(*T); ok {
		return l
	}
	return nil
}
