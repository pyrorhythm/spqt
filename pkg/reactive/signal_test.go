package reactive

import (
	"testing"
)

func TestSignal_Subscribe_fires(t *testing.T) {
	s := NewSignal[int]()
	var got int
	s.Subscribe(func(v int) { got = v })
	s.Emit(99)
	if got != 99 {
		t.Fatalf("expected 99, got %d", got)
	}
}

func TestSignal_Subscribe_unsub(t *testing.T) {
	s := NewSignal[string]()
	count := 0
	unsub := s.Subscribe(func(v string) { count++ })
	s.Emit("hello")
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	unsub()
	s.Emit("world")
	if count != 1 {
		t.Fatalf("expected count still 1 after unsub, got %d", count)
	}
}

func TestSignal_Subscribe_multipleListeners(t *testing.T) {
	s := NewSignal[int]()
	var a, b int
	unsubA := s.Subscribe(func(v int) { a = v })
	s.Subscribe(func(v int) { b = v })
	s.Emit(7)
	if a != 7 || b != 7 {
		t.Fatalf("expected both 7, got a=%d b=%d", a, b)
	}
	unsubA()
	s.Emit(8)
	if a != 7 {
		t.Fatalf("expected a still 7 after unsub, got %d", a)
	}
	if b != 8 {
		t.Fatalf("expected b=8, got %d", b)
	}
}

func TestVoidSignal_Subscribe_fires(t *testing.T) {
	s := NewVoidSignal()
	count := 0
	s.Subscribe(func() { count++ })
	s.Emit()
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestVoidSignal_Subscribe_unsub(t *testing.T) {
	s := NewVoidSignal()
	count := 0
	unsub := s.Subscribe(func() { count++ })
	s.Emit()
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	unsub()
	s.Emit()
	if count != 1 {
		t.Fatalf("expected count still 1 after unsub, got %d", count)
	}
}

func TestVoidSignal_Subscribe_multipleListeners(t *testing.T) {
	s := NewVoidSignal()
	var a, b int
	unsubA := s.Subscribe(func() { a++ })
	s.Subscribe(func() { b++ })
	s.Emit()
	if a != 1 || b != 1 {
		t.Fatalf("expected both 1, got a=%d b=%d", a, b)
	}
	unsubA()
	s.Emit()
	if a != 1 {
		t.Fatalf("expected a still 1 after unsub, got %d", a)
	}
	if b != 2 {
		t.Fatalf("expected b=2, got %d", b)
	}
}
