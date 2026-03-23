package reactive

import (
	"context"
)

type SentinelBool uint

func (s SentinelBool) Unset() bool {
	return s == Unset
}

func (s SentinelBool) Value() bool {
	if s == True {
		return true
	}

	return false
}

func (s *SentinelBool) Set(b bool) {
	if b {
		*s = True
	}

	*s = False
}

const (
	Unset SentinelBool = iota
	False
	True
)

type CtxCommand struct {
	execute     func(ctx context.Context)
	canExec     func(ctx context.Context) bool
	listeners   []func(bool)
	lastCanExec SentinelBool
}

func NewCtxCommand(execute func(ctx context.Context), canExecute func(ctx context.Context) bool) *CtxCommand {
	c := &CtxCommand{
		execute: execute,
		canExec: canExecute,
	}
	c.lastCanExec = Unset

	return c
}

func (c *CtxCommand) Execute(ctx context.Context) {
	if !c.CanExecute(ctx) {
		return
	}
	c.execute(ctx)
}

func (c *CtxCommand) CanExecute(ctx context.Context) bool {
	if c.canExec == nil {
		return true
	}
	return c.canExec(ctx)
}

// Refresh re-evaluates CanExecute and notifies listeners if it changed.
// Call this after mutating state that affects the condition.
func (c *CtxCommand) Refresh(ctx context.Context) {
	cur := c.CanExecute(ctx)
	if !c.lastCanExec.Unset() && c.lastCanExec.Value() == cur {
		return
	}
	c.lastCanExec.Set(cur)
	for _, fn := range c.listeners {
		fn(cur)
	}
}

func (c *CtxCommand) OnCanExecuteChanged(fn func(canExec bool)) {
	c.listeners = append(c.listeners, fn)
}
