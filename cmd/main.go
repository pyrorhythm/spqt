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

	qt.NewQApplication(os.Args)

	aw := qtui.AppWindow{}.Create(ctx)
	aw.MainW.Show()
	aw.AppV.Auth.LoginCmd.Execute(ctx)

	qt.QApplication_Exec()
}
