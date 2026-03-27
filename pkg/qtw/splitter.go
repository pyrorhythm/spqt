package qtw

import (
	qt "github.com/mappu/miqt/qt6"
)

type SplitterBuilder struct {
	s      *qt.QSplitter
	curIdx int
}

func Splitter(orientation qt.Orientation) *SplitterBuilder {
	return &SplitterBuilder{
		s: qt.NewQSplitter3(orientation),
	}
}

func (sb *SplitterBuilder) Widget(w *qt.QWidget, stretchFactor int) *SplitterBuilder {
	sb.s.AddWidget(w)
	sb.s.SetStretchFactor(sb.curIdx, stretchFactor)
	sb.curIdx++

	return sb
}

func (sb *SplitterBuilder) WidgetF(w func() *qt.QWidget, stretchFactor int) *SplitterBuilder {
	sb.s.AddWidget(w())
	sb.s.SetStretchFactor(sb.curIdx, stretchFactor)
	sb.curIdx++

	return sb
}

func (sb *SplitterBuilder) Q() *qt.QSplitter {
	return sb.s
}
