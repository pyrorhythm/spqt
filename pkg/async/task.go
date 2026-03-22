package async

import "github.com/mappu/miqt/qt6/mainthread"

// Run executes work in a goroutine and delivers the result
// to callback on the main (Qt) thread.
func Run[T any](work func() T, callback func(T)) {
	go func() {
		result := work()
		mainthread.Wait(func() {
			callback(result)
		})
	}()
}

// RunErr executes work in a goroutine and delivers (result, error)
// to callback on the main thread.
func RunErr[T any](work func() (T, error), callback func(T, error)) {
	go func() {
		result, err := work()
		mainthread.Wait(func() {
			callback(result, err)
		})
	}()
}

// Fire executes work in a goroutine and calls callback on the main thread
// when done. For side-effect-only background work.
func Fire(work func(), callback func()) {
	go func() {
		work()
		mainthread.Wait(callback)
	}()
}
