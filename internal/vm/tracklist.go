package vm

import (
	"context"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type TrackList struct {
	Tracks  *reactive.List[types.EnrichedTrack]
	Status  *reactive.Prop[string]
	Loading *reactive.Prop[bool]

	LoadCmd *reactive.CtxCommand

	client types.Client
	player *Player
}

func newTrackListVM(player *Player) *TrackList {
	tl := &TrackList{
		Tracks:  reactive.NewList[types.EnrichedTrack](),
		Status:  reactive.NewProp[string](""),
		Loading: reactive.NewProp[bool](false),
		player:  player,
	}

	tl.LoadCmd = reactive.NewCtxCommand(
		func(ctx context.Context) { tl.loadLiked(ctx) },
		func(context.Context) bool { return tl.client != nil && !tl.Loading.Get() },
	)

	return tl
}

func (tl *TrackList) SetClient(c types.Client) {
	tl.client = c
}

func (tl *TrackList) Select(index int) {
	if index >= 0 && index < tl.Tracks.Len() {
		tgt := tl.Tracks.At(index)

		log.Logger().Debug().Any("tgt", tgt).Send()

		tl.player.Current.Set(tgt)
	}
}

func (tl *TrackList) loadLiked(ctx context.Context) {
	tl.Loading.Set(true)
	tl.Status.Set("Loading...")

	go func() {
		cr, err := tl.client.LikedTracks(ctx)
		if err != nil {
			tl.Status.Set(err.Error())
			tl.Loading.Set(false)
			return
		}

		tracks, err := tl.client.EnrichPage(ctx, cr, 0)
		if err != nil {
			tl.Status.Set(err.Error())
			tl.Loading.Set(false)
			return
		}

		tl.Tracks.SetItems(tracks)
		tl.Status.Set("")
		tl.Loading.Set(false)
	}()
}
