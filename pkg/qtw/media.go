package qtw

import (
	qt "github.com/mappu/miqt/qt6"
)

func Placeholder(w, h int) *qt.QPixmap {
	pm := qt.NewQPixmap2(w, h)
	pm.Fill()
	return pm
}

func PlaceholderColor(w, h int, color *qt.QColor) *qt.QPixmap {
	pm := qt.NewQPixmap2(w, h)
	pm.FillWithFillColor(color)
	return pm
}

func PixmapFromBytes(data []byte) (*qt.QPixmap, bool) {
	pm := qt.NewQPixmap()
	ok := pm.LoadFromDataWithData(data)
	return pm, ok
}

func ScaledPixmap(pm *qt.QPixmap, w, h int) *qt.QPixmap {
	return pm.Scaled2(w, h, qt.KeepAspectRatio)
}

func StdIcon(sp qt.QStyle__StandardPixmap) *qt.QIcon {
	return qt.QApplication_Style().StandardIcon(sp, nil, nil)
}

func ThemeIcon(name string) *qt.QIcon {
	return qt.QIcon_FromTheme(name)
}

func ThemeIconOr(name string, fallback *qt.QIcon) *qt.QIcon {
	return qt.QIcon_FromTheme2(name, fallback)
}

func IconPlay() *qt.QIcon  { return StdIcon(qt.QStyle__SP_MediaPlay) }
func IconPause() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaPause) }
func IconStop() *qt.QIcon  { return StdIcon(qt.QStyle__SP_MediaStop) }
func IconNext() *qt.QIcon  { return StdIcon(qt.QStyle__SP_MediaSkipForward) }
func IconPrev() *qt.QIcon  { return StdIcon(qt.QStyle__SP_MediaSkipBackward) }
