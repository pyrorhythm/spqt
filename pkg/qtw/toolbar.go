package qtw

import qt "github.com/mappu/miqt/qt6"

// toolbarSep is a sentinel for toolbar separators.
type toolbarSep struct{}

// ToolSep returns a separator sentinel for use in ToolbarItems.
func ToolSep() toolbarSep { return toolbarSep{} }

// Toolbar creates a QToolBar and populates it.
// Accepts: *qt.QWidget (or any type with .QWidget), toolbarSep, stretchItem.
func Toolbar(items ...any) *qt.QToolBar {
	tb := qt.NewQToolBar3()
	for _, item := range items {
		switch item.(type) {
		case toolbarSep:
			tb.AddSeparator()
		case stretchItem:
			spacer := qt.NewQWidget2()
			spacer.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Preferred)
			tb.AddWidget(spacer)
		default:
			if w := mustWidget(item); w != nil {
				tb.AddWidget(w)
			}
		}
	}
	return tb
}
