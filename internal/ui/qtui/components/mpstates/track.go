package mpstates

import (
	"context"
	"sync/atomic"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

const TrackRowHeight = 72

type trackRowRefs struct {
	cover    *qt.QLabel
	track    *qt.QLabel
	artist   *qt.QLabel
	album    *qt.QLabel
	duration *qt.QLabel
}

func (r *trackRowRefs) yieldItems() []any {
	return []any{
		r.cover,
		qtw.VBox().Margins(4, 4, 4, 4).Items(r.track, r.artist, qtw.StretchN(1), r.album),
		qtw.Stretch(),
		r.duration,
	}
}

func CreateTrackRow() *qt.QWidget {
	w := qtw.Widget()

	refs := &trackRowRefs{
		cover: qtw.EmptyLabel().
			FixedSize(56, 56).ScaledContents(true).
			Pixmap(qtw.PlaceholderColor(56, 56, qt.NewQColor11(99, 99, 99, 255))).
			Align(qt.AlignLeft).Q(),
		track: qtw.
			Label("").Name("trackName").Q(),
		artist: qtw.
			Label("").Name("artistName").
			Fontb(qtw.Font(qtw.DefaultSans).Italic().Size(12)).Q(),
		album: qtw.
			Label("").Name("albumName").
			Fontb(qtw.Font(qtw.DefaultSans).Size(11).Weight(qt.QFont__Light)).Q(),
		duration: qtw.Label("").Align(qt.AlignVCenter).Q(),
	}

	qtw.SetUserData(w.Q(), "trackRefs", refs)

	return w.
		Layout(qtw.HBox().Spacing(6).Margins(4, 4, 4, 4).Items(refs.yieldItems()...)).
		StyleSheet(":hover { background: #2d2d2d }").
		FixedHeight(TrackRowHeight).Q()
}

func BindTrackRow(pl *vm.Player, images *vm.ImageService) func(widget *qt.QWidget, t *metadatapb.Track) {
	return func(widget *qt.QWidget, t *metadatapb.Track) {
		ctx := log.WithCtx(context.Background(), log.Logger().With().Str("track-uri", t.GetCanonicalUri()).Logger())
		rawRefs, ok := qtw.GetUserData(widget, "trackRefs")
		if !ok {
			return
		}
		refs := rawRefs.(*trackRowRefs)
		if t == nil {
			return
		}

		name := t.GetName()
		if name == "" {
			name = "Unknown"
		}
		refs.track.SetText(name)
		refs.artist.SetText(types.ArtistNames(t))
		albumName := "Unknown"
		if t.Album != nil {
			albumName = t.Album.GetName()
		}
		refs.album.SetText(albumName)
		refs.duration.SetText(qtw.FormatDuration(int64(t.GetDuration())))

		refs.cover.SetPixmap(qtw.PlaceholderColor(56, 56, qt.NewQColor11(99, 99, 99, 255)))
		images.LoadCover(t.GetAlbum(), 56, func(pm *qt.QPixmap) {
			if pm != nil {
				refs.cover.SetPixmap(pm)
			}
		})

		var clickLock atomic.Bool
		track := t
		widget.OnEvent(func(f func(*qt.QEvent) bool, e *qt.QEvent) bool {
			if e == nil {
				return false
			}
			switch e.Type() {
			case qt.QEvent__MouseButtonPress:
				if clickLock.CompareAndSwap(false, true) {
					pl.Current.Set(track)
				}
				return true
			case qt.QEvent__MouseButtonRelease:
				clickLock.Store(false)
				return true
			default:
				return f(e)
			}
		})

		log.Trace(ctx).Msg("bound")
	}
}

func PlaceholderTrackRow(widget *qt.QWidget) {
	rawRefs, ok := qtw.GetUserData(widget, "trackRefs")
	if !ok {
		return
	}
	refs := rawRefs.(*trackRowRefs)
	refs.track.SetText("Loading...")
	refs.artist.SetText("")
	refs.album.SetText("")
	refs.duration.SetText("")
	refs.cover.SetPixmap(qtw.PlaceholderColor(56, 56, qt.NewQColor11(55, 55, 55, 255)))
}
