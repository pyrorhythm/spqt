package qtui

import (
	"context"
	"os"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/ui/qtui/components"
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

func CreateAppWindow(
	ctx context.Context,
	auth types.Authenticator,
	clientFactory func(types.Session) types.Client,
) *AppWindow {
	qa := qt.NewQApplication(os.Args)
	aw := &AppWindow{
		MW: qt.NewQMainWindow2(),
		VM: vm.New(ctx, auth, clientFactory),
		qa: qa,
	}

	pages := qtw.NewPages().
		Page("auth", authPage(ctx, aw.VM.Auth)).
		Page("main", components.NewShell(ctx, aw.VM.Shell).Widget())

	aw.VM.Current.OnChange(func(m vm.MetaVM) {
		pages.Show(m.String())
	})

	aw.MW.SetCentralWidget(pages.Widget())

	return aw
}
