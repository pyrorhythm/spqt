package qtw

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type LabelBuilder struct{ lbl *qt.QLabel }

func Label(text string) *LabelBuilder {
	return &LabelBuilder{lbl: qt.NewQLabel3(text)}
}

func EmptyLabel() *LabelBuilder {
	return &LabelBuilder{lbl: qt.NewQLabel2()}
}

func (b *LabelBuilder) Name(n string) *LabelBuilder {
	b.lbl.SetObjectName(*qt.NewQAnyStringView3(n))
	return b
}

func (b *LabelBuilder) Align(a qt.AlignmentFlag) *LabelBuilder {
	b.lbl.SetAlignment(a)
	return b
}

func (b *LabelBuilder) Font(f *qt.QFont) *LabelBuilder {
	b.lbl.SetFont(f)
	return b
}

func (b *LabelBuilder) FixedSize(w, h int) *LabelBuilder {
	b.lbl.SetFixedSize2(w, h)
	return b
}

func (b *LabelBuilder) ScaledContents(on bool) *LabelBuilder {
	b.lbl.SetScaledContents(on)
	return b
}

func (b *LabelBuilder) Pixmap(pm *qt.QPixmap) *LabelBuilder {
	b.lbl.SetPixmap(pm)
	return b
}

func (b *LabelBuilder) WordWrap(on bool) *LabelBuilder {
	b.lbl.SetWordWrap(on)
	return b
}

func (b *LabelBuilder) Property(name string, vart *qt.QVariant) *LabelBuilder {
	b.lbl.SetProperty(name, vart)
	return b
}

func (b *LabelBuilder) BindText(prop *reactive.Prop[string]) *LabelBuilder {
	Bind(prop, b.lbl.SetText)
	return b
}

func (b *LabelBuilder) Build() *qt.QLabel { return b.lbl }

type ButtonBuilder struct{ btn *qt.QPushButton }

func Button(text string) *ButtonBuilder {
	return &ButtonBuilder{btn: qt.NewQPushButton3(text)}
}

func IconButton(icon *qt.QIcon, size int) *ButtonBuilder {
	btn := qt.NewQPushButton2()
	btn.SetIcon(icon)
	btn.SetFixedSize2(size, size)
	return &ButtonBuilder{btn: btn}
}

func (b *ButtonBuilder) Name(n string) *ButtonBuilder {
	b.btn.SetObjectName(*qt.NewQAnyStringView3(n))
	return b
}

func (b *ButtonBuilder) FixedSize(w, h int) *ButtonBuilder {
	b.btn.SetFixedSize2(w, h)
	return b
}

func (b *ButtonBuilder) Icon(icon *qt.QIcon) *ButtonBuilder {
	b.btn.SetIcon(icon)
	return b
}

func (b *ButtonBuilder) Enabled(on bool) *ButtonBuilder {
	b.btn.QWidget.SetEnabled(on)
	return b
}

func (b *ButtonBuilder) Visible(on bool) *ButtonBuilder {
	b.btn.QWidget.SetVisible(on)
	return b
}

func (b *ButtonBuilder) OnClick(fn func()) *ButtonBuilder {
	b.btn.OnClicked(fn)
	return b
}

func (b *ButtonBuilder) BindEnabled(prop *reactive.Prop[bool]) *ButtonBuilder {
	Bind(prop, func(v bool) { b.btn.QWidget.SetEnabled(v) })
	return b
}

func (b *ButtonBuilder) BindIcon(prop *reactive.Prop[bool], whenTrue, whenFalse *qt.QIcon) *ButtonBuilder {
	Bind(prop, func(v bool) {
		if v {
			b.btn.SetIcon(whenTrue)
		} else {
			b.btn.SetIcon(whenFalse)
		}
	})
	return b
}

func (b *ButtonBuilder) Build() *qt.QPushButton { return b.btn }

type SliderBuilder struct{ s *qt.QSlider }

func Slider(orientation qt.Orientation) *SliderBuilder {
	return &SliderBuilder{s: qt.NewQSlider3(orientation)}
}

func HSlider() *SliderBuilder { return Slider(qt.Horizontal) }
func VSlider() *SliderBuilder { return Slider(qt.Vertical) }

func (b *SliderBuilder) Range(min, max int) *SliderBuilder {
	b.s.SetRange(min, max)
	return b
}

func (b *SliderBuilder) Value(v int) *SliderBuilder {
	b.s.SetValue(v)
	return b
}

func (b *SliderBuilder) Name(n string) *SliderBuilder {
	b.s.SetObjectName(*qt.NewQAnyStringView3(n))
	return b
}

func (b *SliderBuilder) BindValue(prop *reactive.Prop[int32]) *SliderBuilder {
	Bind(prop, func(v int32) { b.s.SetValue(int(v)) })
	return b
}

func (b *SliderBuilder) BindRange(min, max *reactive.Prop[int32]) *SliderBuilder {
	Bind(min, func(v int32) { b.s.SetMinimum(int(v)) })
	Bind(max, func(v int32) { b.s.SetMaximum(int(v)) })
	return b
}

func (b *SliderBuilder) Build() *qt.QSlider { return b.s }
