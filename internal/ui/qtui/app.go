package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
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

	pages := qtw.NewPages().
		Page("auth", buildAuthPage(ctx, aw.VM.Auth)).
		Page("player", buildPlayerPage(ctx, aw.VM.Player, aw.VM.TrackList))

	aw.VM.Current.OnChange(func(cvm vm.CurrentViewModel) {
		pages.Show(string(cvm))
	})

	aw.MW.SetCentralWidget(pages.Widget())

	return aw
}
