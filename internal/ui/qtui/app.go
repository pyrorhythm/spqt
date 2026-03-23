package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
)

type AppWindow struct {
	MW *qt.QMainWindow
	VM *vm.App
	qa *qt.QApplication
}

type Themable interface {
	QSS() string
}

func (w AppWindow) SetTheme(t Themable) {
	w.qa.SetStyleSheet(t.QSS())
}

func (AppWindow) Create(ctx context.Context, qa *qt.QApplication, auth types.Authenticator) *AppWindow {
	aw := &AppWindow{
		MW: qt.NewQMainWindow2(),
		VM: vm.New(auth),
		qa: qa,
	}

	pages := qt.NewQStackedWidget2()
	pages.AddWidget(buildAuthPage(ctx, aw.VM.Auth))
	pages.AddWidget(buildPlayerPage(ctx, aw.VM.Player, aw.VM.TrackList))

	aw.VM.Current.OnChange(func(cvm vm.CurrentViewModel) {
		pages.SetCurrentIndex(cvm.Index())
	})

	aw.MW.SetCentralWidget(pages.QWidget)

	return aw
}
