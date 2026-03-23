package qtw

import "github.com/mappu/miqt/qt6/mainthread"

// RunOnMain executes fn on the Qt main thread, blocking until done.
func RunOnMain(fn func()) {
	mainthread.Wait(fn)
}

// RunOnMainAsync executes fn on the Qt main thread without blocking.
func RunOnMainAsync(fn func()) {
	mainthread.Start(fn)
}
