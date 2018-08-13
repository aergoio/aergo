package log

// LazyEval can be used to evaluate an argument under a correct log level.
type LazyEval func() string

func (l LazyEval) String() string {
	return l()
}

// DoLazyEval returns LazyEval. Unnecessary evalution can be prevented by using
// "%v" format string,
func DoLazyEval(c func() string) LazyEval {
	return LazyEval(c)
}
