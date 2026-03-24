package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/respot"
	"github.com/pyrorhythm/spqt/internal/ui/qtui"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	aw := qtui.CreateAppWindow(ctx, respot.Authorize, respot.NewClient)
	aw.MW.SetWindowFlags(qt.FramelessWindowHint)
	aw.MW.Show()
	aw.VM.Auth.LoginCmd.Execute(ctx)
	aw.SetTheme(qtui.AquaDark)
	// aw.MW.SetStyleSheet("QMAborder-radius: 15px;")

	qt.QApplication_Exec()
}
