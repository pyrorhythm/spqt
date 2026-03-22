package qtbind

import "github.com/mappu/miqt/qt6/mainthread"

// OnMainThread executes fn on the Qt main thread, blocking until done.
func OnMainThread(fn func()) {
	mainthread.Wait(fn)
}

// OnMainThreadAsync executes fn on the Qt main thread without blocking.
func OnMainThreadAsync(fn func()) {
	mainthread.Start(fn)
}
