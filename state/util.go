package state

type stack []int

func newStack() *stack {
	return &stack{}
}
func (stk *stack) last() int {
	return len(*stk) - 1
}
func (stk *stack) peek() int {
	if stk == nil || len(*stk) == 0 {
		return -1
	}
	return (*stk)[stk.last()]
}
func (stk *stack) pop() int {
	if stk == nil || len(*stk) == 0 {
		return -1
	}
	lst := stk.last()
	top := (*stk)[lst]
	*stk = (*stk)[:lst]
	return top
}
func (stk *stack) push(args ...int) *stack {
	if stk == nil {
		stk = &stack{}
	}
	*stk = append(*stk, args...)
	return stk
}
func (stk *stack) export() []int {
	size := len(*stk)
	if size <= 0 {
		return nil
	}
	buf := make([]int, size)
	for i, v := range *stk {
		buf[size-i-1] = v
	}
	return buf
}
func (stk *stack) get(index int) int {
	if index < 0 {
		return -1
	}
	return (*stk)[index]
}
func (stk *stack) iter() func() int {
	next := stk.last()
	return func() int {
		if next < 0 {
			return -1
		}
		idx := next
		next--
		return (*stk)[idx]
	}
}
