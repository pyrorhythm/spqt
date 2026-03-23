package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/ui/qtui/components"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func buildPlayerPage(ctx context.Context, pvm *vm.Player, tl *vm.TrackList) *qt.QWidget {
	page := qtw.Widget()
	page.SetLayout(qtw.VBox().NoMargins().Spacing(0).Items(
		components.BuildTrackTable(ctx, tl),
		qtw.HSeparator(),
		components.BuildPlayerBar(pvm),
	))
	return page
}
