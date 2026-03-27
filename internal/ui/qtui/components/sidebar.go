package components

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

func BuildSidebar(nav *reactive.Prop[vm.NavState]) *qt.QWidget {
	return qtw.Widget().Name("sidebar").Layout(
		qtw.VBox().NoMargins().Spacing(0).Items(
			qtw.List().
				Frame(qt.QFrame__NoFrame).
				Item("Home").
				Item("Search").
				Item("Liked Songs").
				OnClick(func(label string) {
					nav.Set(vm.NavStateFrom(label))
				}).
				Stylesheet("border-radius: 0;").
				Margins(8, 24, 8, 24).
				Widget(),
		),
	).Q()
}
