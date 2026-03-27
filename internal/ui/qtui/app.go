package qtui

import (
	"context"
	"os"

	"github.com/dgraph-io/badger/v4"
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
	clientFactory func(context.Context, types.Session) types.Client,
	db *badger.DB,
) *AppWindow {
	qa := qt.NewQApplication(os.Args)
	aw := &AppWindow{
		MW: qt.NewQMainWindow2(),
		VM: vm.New(ctx, auth, clientFactory, db),
		qa: qa,
	}

	pages := qtw.NewPages().
		Page("auth", authPage(ctx, aw.VM.Auth)).
		Page("main", components.NewShell(ctx, aw.VM).Widget())

	qtw.Bind(aw.VM.State, qtw.PageAdapter[vm.AppState](pages))

	aw.MW.SetCentralWidget(pages.Widget())

	return aw
}
