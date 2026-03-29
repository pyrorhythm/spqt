package components

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/ui/qtui/components/mpstates"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

type mainPage struct {
	container *qt.QBoxLayout
	current   qtw.Disposable
	cancelCtx context.CancelFunc
	baseCtx   context.Context
	builders  map[vm.NavState]func(context.Context) qtw.Disposable
}

func (mp *mainPage) widget() *qt.QWidget {
	return qtw.Widget().Layout(mp.container.QLayout).Q()
}

func BuildMainPage(ctx context.Context, app *vm.App) *qt.QWidget {
	ctx = log.Span(ctx, "nav")

	mp := &mainPage{
		container: qtw.VBox().NoMargins().Box(),
		baseCtx:   ctx,
		builders: map[vm.NavState]func(context.Context) qtw.Disposable{
			vm.NavLikedTracks: mpstates.BuildLikedTracks(app),
			vm.NavSearch:      mpstates.BuildSearch(app),
		},
	}

	qtw.Bind(app.Nav, func(state vm.NavState) {
		log.Trace(ctx).Str("page", state.String()).Msg("navigating")
		mp.showPage(state)
	})

	return mp.widget()
}

func (mp *mainPage) showPage(state vm.NavState) {
	if mp.current != nil {
		if mp.cancelCtx != nil {
			mp.cancelCtx()
		}
		mp.current.Dispose()
		mp.current = nil
	}

	for mp.container.Count() > 0 {
		mp.container.TakeAt(0)
	}

	builder, ok := mp.builders[state]
	if !ok {
		lbl := qtw.Label(state.String()).Q()
		mp.container.AddWidget(lbl.QWidget)
		return
	}

	pageCtx, cancel := context.WithCancel(mp.baseCtx)
	mp.cancelCtx = cancel
	mp.current = builder(pageCtx)

	if w, ok := mp.current.(qtw.Widgeter); ok {
		mp.container.AddWidget(w.Widget())
	}
}
