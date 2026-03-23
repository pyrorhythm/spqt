package qtw

import qt "github.com/mappu/miqt/qt6"

// FontBuilder constructs a QFont via chaining.
type FontBuilder struct{ f *qt.QFont }

// Font starts building a QFont with the given family.
func Font(family string) *FontBuilder {
	return &FontBuilder{f: qt.NewQFont2(family)}
}

func (b *FontBuilder) Size(pt int) *FontBuilder {
	b.f.SetPointSize(pt)
	return b
}

func (b *FontBuilder) PixelSize(px int) *FontBuilder {
	b.f.SetPixelSize(px)
	return b
}

func (b *FontBuilder) Bold() *FontBuilder {
	b.f.SetBold(true)
	return b
}

func (b *FontBuilder) Italic() *FontBuilder {
	b.f.SetItalic(true)
	return b
}

func (b *FontBuilder) Weight(w qt.QFont__Weight) *FontBuilder {
	b.f.SetWeight(w)
	return b
}

func (b *FontBuilder) Build() *qt.QFont { return b.f }
