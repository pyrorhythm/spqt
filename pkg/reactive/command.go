package reactive

type Command struct {
	execute    func()
	canExec    func() bool
	listeners  []func(bool)
	lastCanExec bool
}

func NewCommand(execute func(), canExecute func() bool) *Command {
	c := &Command{
		execute: execute,
		canExec: canExecute,
	}
	if canExecute != nil {
		c.lastCanExec = canExecute()
	} else {
		c.lastCanExec = true
	}
	return c
}

func (c *Command) Execute() {
	if !c.CanExecute() {
		return
	}
	c.execute()
}

func (c *Command) CanExecute() bool {
	if c.canExec == nil {
		return true
	}
	return c.canExec()
}

// Refresh re-evaluates CanExecute and notifies listeners if it changed.
// Call this after mutating state that affects the condition.
func (c *Command) Refresh() {
	cur := c.CanExecute()
	if cur == c.lastCanExec {
		return
	}
	c.lastCanExec = cur
	for _, fn := range c.listeners {
		fn(cur)
	}
}

func (c *Command) OnCanExecuteChanged(fn func(canExec bool)) {
	c.listeners = append(c.listeners, fn)
}
