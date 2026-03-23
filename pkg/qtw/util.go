package qtw

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
)

// HSeparator creates a horizontal line separator (QFrame HLine + Sunken).
func HSeparator() *qt.QFrame {
	f := qt.NewQFrame2()
	f.SetFrameShape(qt.QFrame__HLine)
	f.SetFrameShadow(qt.QFrame__Sunken)
	return f
}

// VSeparator creates a vertical line separator.
func VSeparator() *qt.QFrame {
	f := qt.NewQFrame2()
	f.SetFrameShape(qt.QFrame__VLine)
	f.SetFrameShadow(qt.QFrame__Sunken)
	return f
}

// Name sets the object name on any QWidget. Returns the widget for chaining.
func Name(widget *qt.QWidget, name string) *qt.QWidget {
	widget.SetObjectName(*qt.NewQAnyStringView3(name))
	return widget
}

// FormatDuration formats milliseconds as "m:ss".
func FormatDuration(ms int32) string {
	secs := ms / 1000
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

// ScrollArea wraps a widget in a QScrollArea with sensible defaults.
func ScrollArea(widget *qt.QWidget) *qt.QScrollArea {
	sa := qt.NewQScrollArea2()
	sa.SetWidget(widget)
	sa.SetWidgetResizable(true)
	sa.QFrame.SetFrameShape(qt.QFrame__NoFrame)
	return sa
}

// Widget creates an empty QWidget — convenience for container/page construction.
func Widget() *qt.QWidget {
	return qt.NewQWidget2()
}

// WidgetNamed creates an empty QWidget with an object name.
func WidgetNamed(name string) *qt.QWidget {
	w := qt.NewQWidget2()
	w.SetObjectName(*qt.NewQAnyStringView3(name))
	return w
}
