package components

import (
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func BuildPlayerBar(pvm *vm.Player) *qt.QWidget {
	bar := qtw.WidgetNamed("playerBar")
	bar.SetFixedHeight(80)

	coverLbl := qtw.EmptyLabel().FixedSize(56, 56).ScaledContents(true).
		Pixmap(qtw.Placeholder(56, 56)).Align(qt.AlignLeft).Build()

	trackLbl := qtw.Label("—").Name("playerTrackName").Build()
	artistLbl := qtw.EmptyLabel().Name("playerArtistName").Build()

	playBtn := qtw.IconButton(qtw.IconPlay(), 36).FixedSize(36, 36).Build()
	nextBtn := qtw.IconButton(qtw.IconNext(), 36).FixedSize(36, 36).Build()

	posLbl := qtw.Label("0:00").Build()
	slider := qtw.HSlider().Range(0, 1000).Build()
	durLbl := qtw.Label("0:00").Build()

	bar.SetLayout(qtw.HBox().Margins(4, 2, 4, 2).SetStretch(2, 1).Items(
		coverLbl,
		qtw.VBox().Spacing(2).Items(
			trackLbl, artistLbl, qtw.Stretch()),
		qtw.VBox().Spacing(2).Items(
			qtw.HBox().Spacing(12).Items(
				qtw.Stretch(), playBtn, nextBtn, qtw.Stretch()),
			qtw.HBox().Spacing(6).Items(
				posLbl, slider, durLbl),
		),
	))

	// --- Bindings ---
	pvm.Current.OnChange(func(et types.EnrichedTrack) {
		if et.Track == nil {
			trackLbl.SetText("—")
			artistLbl.SetText("")
			durLbl.SetText("0:00")
			slider.SetValue(0)
			coverLbl.SetPixmap(qtw.Placeholder(56, 56))
			return
		}

		trackLbl.SetText(et.Name)

		names := make([]string, 0, len(et.Artists))
		for _, a := range et.Artists {
			names = append(names, a.Name)
		}
		artistLbl.SetText(strings.Join(names, ", "))
		durLbl.SetText(qtw.FormatDuration(et.DurationMs))

		if et.FullAlbum != nil && len(et.FullAlbum.Covers) > 0 {
			qtw.LoadImageAsync(et.FullAlbum.Covers[0].URL, func(pm *qt.QPixmap) {
				coverLbl.SetPixmap(pm)
			}, func(err error) {
				log.Logger().Warn().Err(err).Msg("failed to fetch cover")
			})
		}
	})

	pvm.IsPlaying.OnChange(func(playing bool) {
		if playing {
			playBtn.SetText("⏸")
		} else {
			playBtn.SetText("▶")
		}
	})

	pvm.Progress.OnChange(func(pct float64) {
		slider.SetValue(int(pct * 1000))
		cur := pvm.Current.Get()
		if cur.Track != nil {
			pos := int32(pct * float64(cur.DurationMs) / 1000)
			posLbl.SetText(qtw.FormatDuration(pos))
		}
	})

	if pvm.CanNext != nil {
		pvm.CanNext.OnChange(func(can bool) {
			nextBtn.SetEnabled(can)
		})
	}

	playBtn.OnClicked(func() {
		if pvm.PlayCmd != nil {
			pvm.PlayCmd.Execute()
		}
	})

	nextBtn.OnClicked(func() {
		if pvm.NextCmd != nil {
			pvm.NextCmd.Execute()
		}
	})

	return bar
}
