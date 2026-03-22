package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
)

type AppWindow struct {
	MainW *qt.QMainWindow

	AppV *vm.App
}

func (AppWindow) Create(ctx context.Context) *AppWindow {
	aw := &AppWindow{qt.NewQMainWindow2(), vm.App{}.Create(ctx)}

	pages := qt.NewQStackedWidget2()
	pages.AddWidget(buildAuthPage(ctx, aw.AppV.Auth))

	aw.AppV.Current.OnChange(func(cvm vm.CurrentViewModel) {
		pages.SetCurrentIndex(0)
	})

	aw.MainW.SetCentralWidget(pages.QWidget)

	return aw
}
