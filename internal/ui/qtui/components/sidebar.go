package components

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func BuildSidebar(shell *vm.Shell) *qt.QWidget {
	return qtw.Widget().Name("sidebar").Layout(
		qtw.VBox().NoMargins().Spacing(0).Items(
			qtw.Label("spqt").
				Name("sidebarHeader").
				Property("subheading", qt.NewQVariant14("LIBRARY")).
				Build(),
			qtw.List().
				Frame(qt.QFrame__NoFrame).
				Item("Home").
				Item("Search").
				Item("Liked Songs").
				OnClick(func(label string) {
					shell.CurCtx.Set(vm.SidebarStateFrom(label))
				}).Widget(),
			// qtw.Stretch(),
		),
	).Build()
}
