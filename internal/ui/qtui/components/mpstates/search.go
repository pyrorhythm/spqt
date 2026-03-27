package mpstates

import (
	"context"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

type searchPage struct {
	qtw.Component
	vl    *qtw.VirtualList[*metadatapb.Track]
	query *qt.QLineEdit
}

func (p *searchPage) Dispose() {
	if p.vl != nil {
		p.vl.Destroy()
		p.vl = nil
	}
	if p.query != nil {
		p.query.Delete()
		p.query = nil
	}
	p.Component.Dispose()
}

func (p *searchPage) doSearch(ctx context.Context, app *vm.App) {
	query := p.query.Text()
	if query == "" {
		return
	}

	searchID := app.SearchResults.NewSearchID()
	app.SearchResults.Clear()

	log.Info(ctx).Str("query", query).Msg("searching")

	go func() {
		cr, err := app.Client.Search(ctx, query)
		if err != nil {
			log.Error(ctx).Err(err).Msg("search failed")
			return
		}

		err = app.Client.EnrichPage(ctx, cr, 0, qtw.Guard(func(tracks []*metadatapb.Track) {
			app.SearchResults.AddBatchGuarded(searchID, tracks)
		}))
		if err != nil {
			log.Error(ctx).Err(err).Msg("enrich page failed")
		}
	}()
}

func BuildSearch(app *vm.App) func(context.Context) qtw.Disposable {
	return func(ctx context.Context) qtw.Disposable {
		ctx = log.Span(ctx, "search")

		searchInput := qt.NewQLineEdit2()
		searchInput.SetPlaceholderText("Search tracks...")

		p := &searchPage{
			query: searchInput,
		}

		searchBtn := qtw.Button("Search").Q()
		searchBtn.OnClicked(func() {
			p.doSearch(ctx, app)
		})

		searchRow := qtw.Widget().Layout(
			qtw.HBox().Spacing(8).Items(
				searchInput,
				searchBtn,
			),
		).Q()

		searchInput.OnEditingFinished(func() {
			p.doSearch(ctx, app)
		})

		vl := qtw.NewVirtualList(qtw.VirtualListConfig[*metadatapb.Track]{
			ItemHeight:  TrackRowHeight,
			BufferCount: 10,
			CreateItem:  CreateTrackRow,
			BindItem:    BindTrackRow(app.Player, app.Images),
			Placeholder: PlaceholderTrackRow,
			Data:        app.SearchResults,
		})
		p.vl = vl

		headerLbl := qtw.Label("Search").PropStr("heading", "true").Q()

		return p.Root(
			qtw.Widget().
				Layout(
					qtw.VBox().Margins(24, 24, 24, 24).Spacing(16).Items(
						headerLbl,
						searchRow,
						p.vl,
					),
				).Q())
	}
}
