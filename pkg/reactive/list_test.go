package reactive

import (
	"testing"
)

func TestList_OnChange_fires(t *testing.T) {
	l := NewList[int]()
	count := 0
	l.OnChange(func() { count++ })
	l.Append(1)
	if count != 1 {
		t.Fatalf("expected count 1 after Append, got %d", count)
	}
	l.Append(2)
	if count != 2 {
		t.Fatalf("expected count 2 after second Append, got %d", count)
	}
}

func TestList_OnChange_unsub(t *testing.T) {
	l := NewList[int]()
	count := 0
	unsub := l.OnChange(func() { count++ })
	l.Append(1)
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	unsub()
	l.Append(2)
	if count != 1 {
		t.Fatalf("expected count still 1 after unsub, got %d", count)
	}
}

func TestList_OnChange_multipleListeners(t *testing.T) {
	l := NewList[string]()
	var a, b int
	unsubA := l.OnChange(func() { a++ })
	l.OnChange(func() { b++ })
	l.Append("x")
	if a != 1 || b != 1 {
		t.Fatalf("expected both 1, got a=%d b=%d", a, b)
	}
	unsubA()
	l.Append("y")
	if a != 1 {
		t.Fatalf("expected a still 1 after unsub, got %d", a)
	}
	if b != 2 {
		t.Fatalf("expected b=2, got %d", b)
	}
}

func TestList_Clear_fires(t *testing.T) {
	l := NewList[int]()
	l.Append(1)
	count := 0
	l.OnChange(func() { count++ })
	l.Clear()
	if count != 1 {
		t.Fatalf("expected count 1 after Clear, got %d", count)
	}
}

func TestList_Remove_fires(t *testing.T) {
	l := NewList[int]()
	l.Append(10)
	l.Append(20)
	count := 0
	l.OnChange(func() { count++ })
	l.Remove(0)
	if count != 1 {
		t.Fatalf("expected count 1 after Remove, got %d", count)
	}
	if l.Len() != 1 || l.At(0) != 20 {
		t.Fatalf("unexpected list state after Remove: len=%d", l.Len())
	}
}
