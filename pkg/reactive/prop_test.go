package reactive

import (
	"testing"
)

func TestProp_OnChange_fires(t *testing.T) {
	p := NewProp(0)
	var got int
	p.OnChange(func(v int) { got = v })
	p.Set(42)
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestProp_OnChange_unsub(t *testing.T) {
	p := NewProp(0)
	count := 0
	unsub := p.OnChange(func(v int) { count++ })
	p.Set(1)
	if count != 1 {
		t.Fatalf("expected count 1 after first Set, got %d", count)
	}
	unsub()
	p.Set(2)
	if count != 1 {
		t.Fatalf("expected count still 1 after unsub, got %d", count)
	}
}

func TestProp_OnChange_multipleListeners(t *testing.T) {
	p := NewProp(0)
	var a, b int
	unsubA := p.OnChange(func(v int) { a = v })
	p.OnChange(func(v int) { b = v })
	p.Set(10)
	if a != 10 || b != 10 {
		t.Fatalf("expected both 10, got a=%d b=%d", a, b)
	}
	unsubA()
	p.Set(20)
	if a != 10 {
		t.Fatalf("expected a still 10 after unsub, got %d", a)
	}
	if b != 20 {
		t.Fatalf("expected b=20, got %d", b)
	}
}

func TestProp_OnExact_fires(t *testing.T) {
	p := NewProp(0)
	fired := false
	p.OnExact(5, func() { fired = true })
	p.Set(5)
	if !fired {
		t.Fatal("expected OnExact callback to fire")
	}
}

func TestProp_OnExact_unsub(t *testing.T) {
	p := NewProp(0)
	count := 0
	unsub := p.OnExact(5, func() { count++ })
	p.Set(5)
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	p.Set(0)
	unsub()
	p.Set(5)
	if count != 1 {
		t.Fatalf("expected count still 1 after unsub, got %d", count)
	}
}

func TestProp_SameValue_noFire(t *testing.T) {
	p := NewProp(7)
	count := 0
	p.OnChange(func(int) { count++ })
	p.Set(7)
	if count != 0 {
		t.Fatalf("expected no fire for same value, got count=%d", count)
	}
}
