package reactive

type ECommand[T comparable] struct {
	executables map[T]*Command
}

func NewECommand[T comparable]() *ECommand[T] {
	return &ECommand[T]{executables: make(map[T]*Command)}
}

func (c *ECommand[T]) Register(on T, cmd *Command) {
	c.executables[on] = cmd
}

func (c *ECommand[T]) Execute(val T) {
	if cmd, ok := c.executables[val]; ok {
		cmd.Execute()
	}
}

func (c *ECommand[T]) On(val T) *Command {
	cmd, _ := c.executables[val]
	return cmd
}

//
// func (c *ECommand[T]) Execute() {
// 	if !c.CanExecute() {
// 		return
// 	}
// 	c.execute()
// }
//
// func (c *ECommand[T]) CanExecute() bool {
// 	if c.canExec == nil {
// 		return true
// 	}
// 	return c.canExec()
// }
//
// // Refresh re-evaluates CanExecute and notifies listeners if it changed.
// // Call this after mutating state that affects the condition.
// func (c *ECommand[T]) Refresh() {
// 	cur := c.CanExecute()
// 	if cur == c.lastCanExec {
// 		return
// 	}
// 	c.lastCanExec = cur
// 	for _, fn := range c.listeners {
// 		fn(cur)
// 	}
// }
//
// func (c *ECommand[T]) OnCanExecuteChanged(fn func(canExec bool)) {
// 	c.listeners = append(c.listeners, fn)
// }
