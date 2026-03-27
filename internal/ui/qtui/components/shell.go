package components

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

type Shell struct {
	qtw.Component
}

func NewShell(ctx context.Context, app *vm.App) *qtw.Component {
	c := &Shell{}

	return c.Root(
		qtw.Widget().
			Layout(
				qtw.VBox().
					NoMargins().
					Spacing(0).
					Items(
						qtw.
							Splitter(qt.Horizontal).
							Widget(BuildSidebar(app.Nav), 0).
							Widget(BuildMainPage(ctx, app), 1).Q(),
						qtw.HSeparator(),
						c.Child(NewPlayerBar(ctx, app.Player, app.Images)),
					),
			).Q())
}
