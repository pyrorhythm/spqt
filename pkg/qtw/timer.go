package qtw

import (
	"sync"
	"time"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
)

// Every starts a repeating timer on the Qt main thread.
// Returns a stop function.
func Every(interval time.Duration, fn func()) func() {
	var timer *qt.QTimer

	mainthread.Wait(func() {
		timer = qt.NewQTimer()
		timer.OnTimeout(fn)
		timer.Start(int(interval.Milliseconds()))
	})

	return func() {
		mainthread.Wait(func() {
			timer.Stop()
		})
	}
}

// Debounce returns a function that delays calling fn until after duration
// has elapsed since the last invocation. Safe to call from any goroutine;
// fn runs on the caller's goroutine (use RunOnMain inside fn if needed).
func Debounce(duration time.Duration, fn func()) func() {
	var mu sync.Mutex
	var timer *time.Timer

	return func() {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(duration, fn)
	}
}
