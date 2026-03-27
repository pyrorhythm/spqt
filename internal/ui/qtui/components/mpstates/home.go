package mpstates

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func BuildHome() *qt.QWidget {
	return qtw.Widget().Layout(
		qtw.VBox().Margins(24, 24, 24, 24).Spacing(16).Items(
			qtw.Label("Good afternoon").Property("heading", qt.NewQVariant14("true")).Q(),
			qtw.Stretch(),
		)).Q()
}
