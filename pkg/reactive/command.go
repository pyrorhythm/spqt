package reactive

import (
	"sync"
)

type Command struct {
	mu          sync.RWMutex
	execute     func()
	canExec     func() bool
	nextID      uint64
	listeners   map[uint64]func(bool)
	lastCanExec bool
}

func NewCommand(execute func(), canExecute func() bool) *Command {
	c := &Command{
		execute:   execute,
		canExec:   canExecute,
		listeners: make(map[uint64]func(bool)),
	}
	if canExecute != nil {
		c.lastCanExec = canExecute()
	} else {
		c.lastCanExec = true
	}
	return c
}

func (c *Command) Execute() {
	c.mu.RLock()
	exec := c.execute
	canExec := c.canExec
	c.mu.RUnlock()

	if canExec != nil && !canExec() {
		return
	}
	exec()
}

func (c *Command) CanExecute() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.canExec == nil {
		return true
	}
	return c.canExec()
}

// Refresh re-evaluates CanExecute and notifies listeners if it changed.
func (c *Command) Refresh() {
	c.mu.Lock()
	cur := c.CanExecute()
	if cur == c.lastCanExec {
		c.mu.Unlock()
		return
	}
	c.lastCanExec = cur
	fns := make([]func(bool), 0, len(c.listeners))
	for _, fn := range c.listeners {
		fns = append(fns, fn)
	}
	c.mu.Unlock()
	for _, fn := range fns {
		fn(cur)
	}
}

func (c *Command) OnCanExecuteChanged(fn func(canExec bool)) func() {
	id := nextIDUint64()
	c.mu.Lock()
	c.listeners[id] = fn
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		delete(c.listeners, id)
		c.mu.Unlock()
	}
}
