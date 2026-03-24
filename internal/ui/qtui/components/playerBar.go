package components

import (
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type PlayerBar struct {
	qtw.Component
}

func NewPlayerBar(player *vm.Player) *PlayerBar {
	c := &PlayerBar{}

	// Derived props — split EnrichedTrack into bindable atoms
	durationMs := reactive.Derive(player.Current, func(et types.EnrichedTrack) int32 {
		if et.Track == nil {
			return 0
		}
		return et.DurationMs
	})
	// Widgets — bindings inline
	coverLbl := qtw.EmptyLabel().
		FixedSize(56, 56).
		ScaledContents(true).
		Pixmap(qtw.Placeholder(56, 56)).
		Align(qt.AlignLeft).
		Build()

	trackLbl := qtw.Label("—").
		Name("playerTrackName").
		BindText(reactive.Derive(player.Current,
			func(et types.EnrichedTrack) string {
				if et.Track == nil {
					return "—"
				}
				return et.Name
			})).Build()

	artistLbl := qtw.EmptyLabel().
		Name("playerArtistName").
		BindText(reactive.Derive(player.Current,
			func(et types.EnrichedTrack) string {
				if et.Track == nil {
					return ""
				}
				names := make([]string, 0, len(et.Artists))
				for _, a := range et.Artists {
					names = append(names, a.Name)
				}
				return strings.Join(names, ", ")
			})).Build()

	prevBtn := qtw.
		IconButton(qtw.IconPrev(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCPrev)).
		Build()

	playBtn := qtw.
		IconButton(qtw.IconPlay(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCPlayPause)).
		BindIcon(player.IsPlaying, qtw.IconPause(), qtw.IconPlay()).
		Build()

	nextBtn := qtw.
		IconButton(qtw.IconNext(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCNext)).
		BindEnabled(player.CanNext).
		Build()

	posLbl := qtw.
		Label("0:00").
		BindText(reactive.Derive(player.Progress, qtw.FormatDuration)).
		Build()

	slider := qtw.HSlider().
		Range(0, 0).
		BindValue(player.Progress).
		BindRange(reactive.Always[int32](0), durationMs).
		Build()

	durLbl := qtw.
		Label("0:00").
		BindText(reactive.Derive(durationMs, qtw.FormatDuration)).
		Build()

	qtw.Watch(player.Current, func(et types.EnrichedTrack) {
		if et.Track == nil {
			coverLbl.SetPixmap(qtw.Placeholder(56, 56))
			return
		}
		if et.FullAlbum != nil && len(et.FullAlbum.Covers) > 0 {
			qtw.LoadImageAsync(et.FullAlbum.Covers[0].URL, func(pm *qt.QPixmap) {
				coverLbl.SetPixmap(pm)
			}, func(err error) {
				log.Logger().Warn().Err(err).Msg("failed to fetch cover")
			})
		}
	})

	c.Root(
		qtw.Widget().
			Name("playerBar").
			FixedHeight(80).
			Layout(qtw.HBox().Margins(4, 2, 4, 2).
				SetStretch(2, 1).Items(
				coverLbl,
				qtw.VBox().Spacing(2).Items(trackLbl, artistLbl, qtw.Stretch()),
				qtw.VBox().Spacing(2).Items(
					qtw.HBox().Spacing(12).Items(prevBtn, playBtn, nextBtn, qtw.Stretch()),
					qtw.HBox().Spacing(6).Items(posLbl, slider, durLbl),
				),
			)).Build(),
	)

	return c
}
