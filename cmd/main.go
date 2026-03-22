package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/ui/qtui"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	aw := qtui.AppWindow{}.Create(ctx, qt.NewQApplication(os.Args))
	aw.MainW.Show()
	aw.AppV.Auth.LoginCmd.Execute(ctx)
	aw.SetTheme(qtui.AquaDark)

	qt.QApplication_Exec()
}
