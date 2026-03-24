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

func NewShell(ctx context.Context, shell *vm.Shell) *Shell {
	c := &Shell{}

	c.Root(qtw.Widget().
		Layout(
			qtw.VBox().NoMargins().Spacing(0).Items(
				qtw.
					Splitter(qt.Horizontal).
					Widget(BuildSidebar(shell), 0).
					Widget(BuildMainPage(ctx, shell), 1).
					Build(),
				qtw.HSeparator(),
				c.Child(NewPlayerBar(shell.Player)),
			),
		).Build())

	return c
}
