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

type WidgetBuilder struct{ w *qt.QWidget }

func Widget() *WidgetBuilder {
	return &WidgetBuilder{w: qt.NewQWidget2()}
}

func (b *WidgetBuilder) Name(n string) *WidgetBuilder {
	b.w.SetObjectName(*qt.NewQAnyStringView3(n))
	return b
}

func (b *WidgetBuilder) FixedSize(w, h int) *WidgetBuilder {
	b.w.SetFixedSize2(w, h)
	return b
}

func (b *WidgetBuilder) MinSize(w, h int) *WidgetBuilder {
	b.w.SetMinimumSize2(w, h)
	return b
}

func (b *WidgetBuilder) MaxSize(w, h int) *WidgetBuilder {
	b.w.SetMaximumSize2(w, h)
	return b
}

func (b *WidgetBuilder) Layout(l *qt.QLayout) *WidgetBuilder {
	b.w.SetLayout(l)
	return b
}

func (b *WidgetBuilder) Visible(on bool) *WidgetBuilder {
	b.w.SetVisible(on)
	return b
}

func (b *WidgetBuilder) FixedHeight(h int) *WidgetBuilder {
	b.w.SetFixedHeight(h)
	return b
}

func (b *WidgetBuilder) FixedWidth(w int) *WidgetBuilder {
	b.w.SetFixedWidth(w)
	return b
}

func (b *WidgetBuilder) Build() *qt.QWidget { return b.w }
