package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"runtime/trace"
	"syscall"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/respot"
	"github.com/pyrorhythm/spqt/internal/ui/qtui"
	"github.com/pyrorhythm/spqt/pkg/cache"
	"github.com/pyrorhythm/spqt/pkg/log"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatal(ctx).Msgf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(ctx).Msgf("failed to close trace file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatal(ctx).Msgf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	bdg, err := cache.NewBadger()
	if err != nil {
		log.Logger().Fatal().Err(err).Send()
	}

	aw := qtui.CreateAppWindow(ctx, respot.Authorize, respot.NewClient, bdg)
	aw.MW.SetWindowFlags(qt.FramelessWindowHint)
	aw.MW.Show()
	aw.VM.Auth.LoginCmd.Execute(ctx)
	aw.SetTheme(qtui.AquaDark)

	qt.QApplication_Exec()
}
