package qtw

// DataSource is a generic read interface over an ordered collection
// that supports synchronous hits, async loading, and length-change
// notifications.  VirtualList consumes this interface.
type DataSource[T any] interface {
	// Len returns the current number of items.
	Len() int

	// Get returns the item at index if it is already in-memory.
	// Returns nil, false when the item must be fetched asynchronously.
	Get(index int) (*T, bool)

	// LoadAsync fetches the item at index in the background and invokes cb
	// with the result once available.  cb may be called on any goroutine;
	// callers are responsible for marshalling to the Qt main thread if needed.
	LoadAsync(index int, cb func(T))

	// OnLengthChanged registers a callback that is invoked whenever the
	// total item count changes.  Returns an unsubscribe function.
	OnLengthChanged(func(int)) func()
}
