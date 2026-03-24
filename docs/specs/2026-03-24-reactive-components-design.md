# Reactive Component Model

## Goal

Bring React-like declarative reactivity to Go/Qt components. State-to-widget bindings live inline with widget construction. Components are composable first-class values with lifecycle hooks (disposal deferred).

## Design: Approach 2 + Derived

### New Reactive Primitives (`pkg/reactive/`)

**`Observable[T]` interface** — unifies `Prop[T]` and `CmpProp[T,K]`:

```go
type Observable[T any] interface {
    Get() T
    OnChange(func(T))
}
```

Both `Prop` and `CmpProp` already satisfy this. No changes to existing types.

**`Derive[S,T]`** — computed prop from any observable:

```go
func Derive[S any, T comparable](source Observable[S], fn func(S) T) *Prop[T]
```

Returns a plain `*Prop[T]` (works with all existing `Bind*` methods). Source fires → fn runs → result Set on derived prop (deduped).

**`Always[T]`** — constant prop that never changes:

```go
func Always[T comparable](v T) *Prop[T]
```

Sugar for `NewProp(v)`. Useful as static arg to `BindRange`.

### Component Model (`pkg/qtw/`)

**`Widgeter` interface:**

```go
type Widgeter interface {
    Widget() *qt.QWidget
}
```

**`Component` struct:**

```go
type Component struct {
    root     *qt.QWidget
    children []Widgeter
}
```

Methods: `SetRoot(*qt.QWidget)`, `Widget() *qt.QWidget`, `Child(Widgeter) Widgeter`.

**`Watch` function** (free function — Go can't have generic methods on non-generic types):

```go
func Watch[T any](c *Component, obs reactive.Observable[T], fn func(T))
```

Calls fn immediately with current value, subscribes to changes. Becomes disposal hook later.

### Builder Binding Extensions (`pkg/qtw/widget.go`)

- `LabelBuilder.BindText(prop *reactive.Prop[string])`
- `SliderBuilder.BindValue(prop *reactive.Prop[int32])`
- `SliderBuilder.BindRange(min, max *reactive.Prop[int32])`

Only add methods as actually needed. BindEnabled/BindIcon on ButtonBuilder already exist.

### Layout Integration (`pkg/qtw/layout.go`)

Add `Widgeter` case to `Items()` type-switch in BoxBuilder and GridBuilder:

```go
case Widgeter:
    layout.AddWidget(v.Widget())
```

### WidgetBuilder Additions (`pkg/qtw/util.go`)

- `FixedHeight(h int)`
- `FixedWidth(w int)`

### Migration: playerBar

Before: 4 separate OnChange blocks (25 lines of imperative wiring).
After: Derive splits EnrichedTrack into atoms (trackName, artistName, durationMs, posText, durText). All bind inline. Only async image load remains as Watch.

## Files Changed

| File | Change |
|---|---|
| `pkg/reactive/observable.go` | NEW: Observable interface |
| `pkg/reactive/derived.go` | NEW: Derive, Always |
| `pkg/qtw/component.go` | NEW: Component, Widgeter, Watch |
| `pkg/qtw/widget.go` | ADD: BindText, BindValue, BindRange |
| `pkg/qtw/util.go` | ADD: FixedHeight, FixedWidth |
| `pkg/qtw/layout.go` | MOD: Widgeter case in Items() |
| `internal/ui/qtui/components/playerBar.go` | MIGRATE: to component pattern |
| `internal/ui/qtui/components/shell.go` | MIGRATE: to component pattern |

## Constraints

- Disposal deferred — no unsubscribe mechanism yet
- All Qt ops on main thread (existing RunOnMain handles this)
- Derived returns *Prop[T] requiring T comparable
- Watch is a free function due to Go generics limitation
