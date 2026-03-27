package qtw

import "github.com/mappu/miqt/qt6/mainthread"

// RunOnMain executes fn on the Qt main thread, blocking until done.
func RunOnMain(fn func()) {
	mainthread.Wait(fn)
}

func Guard[A any](fn func(A)) func(A) {
	return func(a A) {
		mainthread.Wait(func() {
			fn(a)
		})
	}
}

func G(fn func()) func() {
	return func() {
		mainthread.Wait(fn)
	}
}

func T[A any](fn func(A), a A) {
	mainthread.Wait(func() {
		fn(a)
	})
}

func T2[A, B any](fn func(A, B), a A, b B) {
	mainthread.Wait(func() {
		fn(a, b)
	})
}

func T3[A, B, C any](fn func(A, B, C), a A, b B, c C) {
	mainthread.Wait(func() {
		fn(a, b, c)
	})
}

// RunOnMainAsync executes fn on the Qt main thread without blocking.
func RunOnMainAsync(fn func()) {
	mainthread.Start(fn)
}
