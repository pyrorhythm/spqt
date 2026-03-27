package mpstates

import (
	"context"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

type likedPage struct {
	qtw.Component
	vl *qtw.VirtualList[*metadatapb.Track]
}

func (p *likedPage) Dispose() {
	if p.vl != nil {
		p.vl.Destroy()
		p.vl = nil
	}
	p.Component.Dispose()
}

func BuildLikedTracks(app *vm.App) func(context.Context) qtw.Disposable {
	return func(ctx context.Context) qtw.Disposable {
		ctx = log.Span(ctx, "liked")
		log.Trace(ctx).Msg("building virtual track list")

		go func() {
			err := app.Client.FetchLikesEnrich(ctx, qtw.Guard(app.LikedTracks.AddBatch))
			if err != nil {
				log.Error(ctx).Err(err).Msg("FetchLikesEnrich failed")
			}
		}()

		vl := qtw.NewVirtualList(qtw.VirtualListConfig[*metadatapb.Track]{
			ItemHeight:  TrackRowHeight,
			BufferCount: 10,
			CreateItem:  CreateTrackRow,
			BindItem:    BindTrackRow(app.Player, app.Images),
			Placeholder: PlaceholderTrackRow,
			Data:        app.LikedTracks,
		})

		c := &likedPage{vl: vl}

		return c.Root(qtw.Widget().
			Layout(
				qtw.VBox().Margins(24, 24, 24, 24).Spacing(16).Items(
					qtw.Label("Liked Songs").PropStr("heading", "true").Q(),
					c.vl,
				),
			).Q())
	}
}
