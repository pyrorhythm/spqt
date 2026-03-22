package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
)

type AppWindow struct {
	MainW *qt.QMainWindow

	AppV *vm.App

	qtapp *qt.QApplication
}

type Themable interface {
	QSS() string
}

func (w AppWindow) SetTheme(t Themable) {
	w.qtapp.SetStyleSheet(t.QSS())
}

func (AppWindow) Create(ctx context.Context, qa *qt.QApplication) *AppWindow {
	aw := &AppWindow{
		MainW: qt.NewQMainWindow2(),
		AppV:  vm.App{}.Create(ctx),
		qtapp: qa,
	}

	pages := qt.NewQStackedWidget2()
	pages.AddWidget(buildAuthPage(ctx, aw.AppV.Auth))

	aw.AppV.Current.OnChange(func(cvm vm.CurrentViewModel) {
		pages.SetCurrentIndex(0)
	})

	aw.MainW.SetCentralWidget(pages.QWidget)

	return aw
}
