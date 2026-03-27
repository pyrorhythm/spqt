package components

import (
	"context"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type PlayerBar struct {
	qtw.Component
	unsubs []func()
}

func (p *PlayerBar) Dispose() {
	for _, unsub := range p.unsubs {
		unsub()
	}
	p.unsubs = nil
	p.Component.Dispose()
}

func NewPlayerBar(ctx context.Context, player *vm.Player, images *vm.ImageService) *PlayerBar {
	c := &PlayerBar{unsubs: make([]func(), 0, 6)}

	placeholder := qtw.PlaceholderColor(56, 56, qt.NewQColor11(0, 0, 0, 0))

	durationMs, unsub := reactive.
		Derive(reactive.Observable[*metadatapb.Track](player.Current),
			func(pb *metadatapb.Track) int64 {
				if pb == nil {
					return 0
				}
				return int64(pb.GetDuration())
			})
	c.unsubs = append(c.unsubs, unsub)

	coverLbl := qtw.EmptyLabel().
		FixedSize(56, 56).
		ScaledContents(true).
		Pixmap(placeholder).
		Align(qt.AlignLeft).Q()

	trackNameProp, unsub := reactive.Derive(reactive.Observable[*metadatapb.Track](player.Current),
		func(pb *metadatapb.Track) string {
			if pb == nil {
				return "—"
			}
			return pb.GetName()
		})
	c.unsubs = append(c.unsubs, unsub)
	trackLbl := qtw.Label("—").
		Name("playerTrackName").
		BindText(trackNameProp).Q()

	artistNameProp, unsub := reactive.Derive(reactive.Observable[*metadatapb.Track](player.Current),
		func(pb *metadatapb.Track) string {
			return types.ArtistNames(pb)
		})
	c.unsubs = append(c.unsubs, unsub)
	artistLbl := qtw.EmptyLabel().
		Name("playerArtistName").
		BindText(artistNameProp).Q()

	prevBtn := qtw.
		IconButton(qtw.IconPrev(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCPrev)).
		Q()

	playBtn := qtw.
		IconButton(qtw.IconPlay(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCPlayPause)).
		BindIcon(player.IsPlaying, qtw.IconPause(), qtw.IconPlay()).
		Q()

	nextBtn := qtw.
		IconButton(qtw.IconNext(), 36).
		FixedSize(36, 36).
		OnClick(player.Exec(vm.PCNext)).
		BindEnabled(player.CanNext).
		Q()

	posProp, unsub := reactive.Derive(reactive.Observable[int64](player.Progress), qtw.FormatDuration)
	c.unsubs = append(c.unsubs, unsub)
	posLbl := qtw.
		Label("0:00").
		BindText(posProp).
		Q()

	slider := qtw.HSlider().
		Range(0, 0).
		BindSeekable(player.Progress, player.SeekPos)
	unsubMin, unsubMax := slider.BindRange(reactive.Always[int64](0), durationMs)
	c.unsubs = append(c.unsubs, unsubMin, unsubMax)
	_ = slider.Q()

	durProp, unsub := reactive.Derive(reactive.Observable[int64](durationMs), qtw.FormatDuration)
	c.unsubs = append(c.unsubs, unsub)
	durLbl := qtw.
		Label("0:00").
		BindText(durProp).
		Q()

	// Lazy album fetch for cover art
	unsub = qtw.Watch(player.Current, func(pb *metadatapb.Track) {
		if pb == nil {
			coverLbl.SetPixmap(placeholder)
			return
		}
		images.LoadCover(pb.GetAlbum(), 56, func(pm *qt.QPixmap) {
			if pm != nil {
				coverLbl.SetPixmap(pm)
			}
		})
	})
	c.unsubs = append(c.unsubs, unsub)

	c.Root(
		qtw.Widget().
			Name("playerBar").
			FixedHeight(80).
			Layout(
				qtw.HBox().
					Margins(4, 8, 4, 8).
					Stretch(2, 1).
					Items(
						coverLbl,
						qtw.VBox().Spacing(2).Items(
							qtw.HBox().Spacing(12).Items(
								prevBtn, playBtn, nextBtn,
								qtw.Stretch(),
								qtw.VBox().Spacing(2).Items(trackLbl, artistLbl),
							),
							qtw.HBox().Spacing(6).Items(posLbl, slider, durLbl),
						),
					)).Q(),
	)

	return c
}
