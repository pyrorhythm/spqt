package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/ui/qtui/components"
	"github.com/pyrorhythm/spqt/internal/vm"
)

func buildPlayerPage(ctx context.Context, pvm *vm.Player, tl *vm.TrackList) *qt.QWidget {
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout2()
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	page.SetLayout(layout.QLayout)

	layout.AddWidget(components.BuildTrackTable(ctx, tl))

	sep := qt.NewQFrame2()
	sep.SetFrameShape(qt.QFrame__HLine)
	sep.SetFrameShadow(qt.QFrame__Sunken)
	layout.AddWidget(sep.QWidget)

	layout.AddWidget(components.BuildPlayerBar(pvm))

	return page
}
