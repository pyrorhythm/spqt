package qtw

import (
	"sync"

	"github.com/mappu/miqt/qt6/mainthread"

	qt "github.com/mappu/miqt/qt6"
)

// VirtualListConfig holds all configuration for a VirtualList.
type VirtualListConfig[T any] struct {
	// ItemHeight is the fixed pixel height of every row.
	ItemHeight int

	// BufferCount is the number of extra rows rendered above and below the
	// visible viewport (default 10).
	BufferCount int

	// CreateItem is called once per pool slot to construct the recycled widget.
	CreateItem func() *qt.QWidget

	// BindItem populates widget with the item data for a cache-hit row.
	BindItem func(widget *qt.QWidget, item T)

	// Placeholder is called when the item is not yet in the cache so the
	// widget can show a loading state while the async fetch is in flight.
	Placeholder func(widget *qt.QWidget)

	// Data is the backing DataSource.
	Data DataSource[T]
}

type poolSlot struct {
	widget *qt.QWidget
	index  int
}

type VirtualList[T any] struct {
	cfg     VirtualListConfig[T]
	pool    []poolSlot
	mu      sync.Mutex
	scrollArea *qt.QScrollArea
	content    *qt.QWidget
	data       DataSource[T]
	unsub      func()
	poolSize   int
}

// NewVirtualList constructs and returns a VirtualList.
//
// The returned widget is not yet attached to any parent; callers should use
// VirtualList.Widget() and add it to their layout.
func NewVirtualList[T any](cfg VirtualListConfig[T]) *VirtualList[T] {
	if cfg.BufferCount == 0 {
		cfg.BufferCount = 10
	}

	vl := &VirtualList[T]{
		cfg:  cfg,
		data: cfg.Data,
	}

	vl.content = qt.NewQWidget2()
	vl.content.SetMinimumHeight(cfg.Data.Len() * cfg.ItemHeight)

	vl.scrollArea = qt.NewQScrollArea2()
	vl.scrollArea.SetWidget(vl.content)
	vl.scrollArea.SetWidgetResizable(true)
	vl.scrollArea.SetFrameShape(qt.QFrame__NoFrame)

	initialPoolSize := 1 + 2*cfg.BufferCount + 20
	vl.poolSize = initialPoolSize
	vl.pool = make([]poolSlot, initialPoolSize)
	for i := range vl.pool {
		w := cfg.CreateItem()
		w.SetParent(vl.content)
		w.SetFixedHeight(cfg.ItemHeight)
		w.SetVisible(false)
		vl.pool[i] = poolSlot{widget: w, index: -1}
	}

	vl.unsub = cfg.Data.OnLengthChanged(func(n int) {
		vl.content.SetMinimumHeight(n * cfg.ItemHeight)
		vl.rebind()
	})

	vl.scrollArea.VerticalScrollBar().OnValueChanged(func(_ int) {
		vl.rebind()
	})

	// Re-layout pool widgets when viewport resizes (e.g. window resize).
	vl.scrollArea.OnResizeEvent(func(super func(*qt.QResizeEvent), event *qt.QResizeEvent) {
		super(event)
		vl.relayout()
	})

	vl.rebind()

	return vl
}

// Widget returns the outer QWidget (the scroll area) for layout embedding.
func (vl *VirtualList[T]) Widget() *qt.QWidget {
	return vl.scrollArea.QWidget
}

// ScrollY returns the current vertical scroll position in pixels.
func (vl *VirtualList[T]) ScrollY() int {
	return vl.scrollArea.QAbstractScrollArea.VerticalScrollBar().Value()
}

// SetScrollY scrolls to the given pixel offset.
func (vl *VirtualList[T]) SetScrollY(y int) {
	vl.scrollArea.QAbstractScrollArea.VerticalScrollBar().SetValue(y)
}

// Destroy unsubscribes all listeners and deletes all pool widgets.
// Call this when the VirtualList is removed from the UI.
func (vl *VirtualList[T]) Destroy() {
	if vl.unsub != nil {
		vl.unsub()
		vl.unsub = nil
	}
	for i := range vl.pool {
		if vl.pool[i].widget != nil {
			DeleteUserData(vl.pool[i].widget)
			vl.pool[i].widget.Delete()
			vl.pool[i].widget = nil
		}
	}
}

// growPool ensures the pool has at least n slots, growing if necessary.
func (vl *VirtualList[T]) growPool(n int) {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	if n <= len(vl.pool) {
		return
	}
	for len(vl.pool) < n {
		w := vl.cfg.CreateItem()
		w.SetParent(vl.content)
		w.SetFixedHeight(vl.cfg.ItemHeight)
		w.SetVisible(false)
		vl.pool = append(vl.pool, poolSlot{widget: w, index: -1})
	}
}

// rebind repositions pool slots to cover the visible region.
// Slots already showing the correct data index are skipped entirely.
func (vl *VirtualList[T]) rebind() {
	total := vl.data.Len()

	firstIdx := max(vl.ScrollY()/vl.cfg.ItemHeight-vl.cfg.BufferCount, 0)
	vpWidth := vl.scrollArea.Viewport().Width()

	// Ensure pool is large enough for the current view.
	requiredSlots := vl.scrollArea.Viewport().Height()/vl.cfg.ItemHeight + 2*vl.cfg.BufferCount + 1
	if firstIdx+requiredSlots > len(vl.pool) {
		vl.growPool(firstIdx + requiredSlots)
	}

	for slotIdx := range vl.pool {
		slot := &vl.pool[slotIdx]
		dataIdx := firstIdx + slotIdx

		if dataIdx >= total {
			// Beyond list end — hide if currently showing something.
			if slot.index != -1 {
				slot.widget.SetVisible(false)
				slot.index = -1
			}
			continue
		}

		if slot.widget == nil {
			continue
		}

		// Already bound to correct data — nothing to do.
		if slot.index == dataIdx {
			continue
		}

		// Recycle this slot for a new data index.
		slot.widget.Move(0, dataIdx*vl.cfg.ItemHeight)
		if vpWidth > 0 {
			slot.widget.Resize(vpWidth, vl.cfg.ItemHeight)
		}
		slot.widget.SetVisible(true)
		slot.index = dataIdx

		if item, ok := vl.data.Get(dataIdx); ok {
			vl.cfg.BindItem(slot.widget, *item)
		} else {
			if vl.cfg.Placeholder != nil {
				vl.cfg.Placeholder(slot.widget)
			}
			capturedIdx := dataIdx
			capturedSlot := slotIdx
			vl.data.LoadAsync(dataIdx, func(item T) {
				mainthread.Wait(func() {
					vl.mu.Lock()
					stillValid := vl.pool[capturedSlot].index == capturedIdx
					vl.mu.Unlock()
					if stillValid {
						vl.cfg.BindItem(vl.pool[capturedSlot].widget, item)
					}
					// No full rebind needed — rebind is only triggered by scroll or length change.
				})
			})
		}
	}
}

// relayout repositions and resizes all visible slots without rebinding data.
// Called on viewport resize.
func (vl *VirtualList[T]) relayout() {
	itemH := vl.cfg.ItemHeight
	vpWidth := vl.scrollArea.QAbstractScrollArea.Viewport().Width()
	if vpWidth <= 0 {
		return
	}
	vl.mu.Lock()
	defer vl.mu.Unlock()
	for slotIdx := range vl.pool {
		slot := &vl.pool[slotIdx]
		if slot.index < 0 || slot.widget == nil {
			continue
		}
		slot.widget.Move(0, slot.index*itemH)
		slot.widget.Resize(vpWidth, itemH)
	}
}
