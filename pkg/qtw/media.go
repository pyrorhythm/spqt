package qtw

import (
	"io"
	"net/http"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

// Placeholder creates a solid-filled pixmap of the given size.
func Placeholder(w, h int) *qt.QPixmap {
	pm := qt.NewQPixmap2(w, h)
	pm.Fill()
	return pm
}

// PlaceholderColor creates a pixmap filled with the given color.
func PlaceholderColor(w, h int, color *qt.QColor) *qt.QPixmap {
	pm := qt.NewQPixmap2(w, h)
	pm.FillWithFillColor(color)
	return pm
}

// PixmapFromBytes loads a pixmap from raw image data.
func PixmapFromBytes(data []byte) (*qt.QPixmap, bool) {
	pm := qt.NewQPixmap()
	ok := pm.LoadFromDataWithData(data)
	return pm, ok
}

// ScaledPixmap scales a pixmap to the given size, keeping aspect ratio.
func ScaledPixmap(pm *qt.QPixmap, w, h int) *qt.QPixmap {
	return pm.Scaled2(w, h, qt.KeepAspectRatio)
}

// --- Icon helpers ---

// StdIcon returns a standard icon from the current style.
func StdIcon(sp qt.QStyle__StandardPixmap) *qt.QIcon {
	return qt.QApplication_Style().StandardIcon(sp, nil, nil)
}

// ThemeIcon returns an icon from the current icon theme.
func ThemeIcon(name string) *qt.QIcon {
	return qt.QIcon_FromTheme(name)
}

// ThemeIconOr returns a theme icon with a fallback.
func ThemeIconOr(name string, fallback *qt.QIcon) *qt.QIcon {
	return qt.QIcon_FromTheme2(name, fallback)
}

// IconPlay returns the platform media-play icon.
func IconPlay() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaPlay) }

// IconPause returns the platform media-pause icon.
func IconPause() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaPause) }

// IconStop returns the platform media-stop icon.
func IconStop() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaStop) }

// IconNext returns the platform media-skip-forward icon.
func IconNext() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaSkipForward) }

// IconPrev returns the platform media-skip-backward icon.
func IconPrev() *qt.QIcon { return StdIcon(qt.QStyle__SP_MediaSkipBackward) }

// LoadImageAsync fetches an image from url in a goroutine, then calls onLoad
// on the main thread with the resulting QPixmap. onErr is called on the
// goroutine if the fetch fails (may be nil to ignore errors).
func LoadImageAsync(url string, onLoad func(*qt.QPixmap), onErr func(error)) {
	go func() {
		data, err := fetchImage(url)
		if err != nil {
			if onErr != nil {
				onErr(err)
			}
			return
		}

		mainthread.Wait(func() {
			pm, ok := PixmapFromBytes(data)
			if ok {
				onLoad(pm)
			}
		})
	}()
}

func fetchImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
