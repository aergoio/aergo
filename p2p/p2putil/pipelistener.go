package p2putil

// MultiListener can contain multiple unit listeners and toss events
type MultiListener struct {
	ls []PipeEventListener
}

func NewMultiListener(ls ...PipeEventListener) *MultiListener {
	return &MultiListener{ls: ls}
}

func (ml *MultiListener) AppendListener(l PipeEventListener) {
	ml.ls = append(ml.ls, l)
}
func (ml *MultiListener) OnIn(element interface{}) {
	for _, l := range ml.ls {
		l.OnIn(element)
	}
}

func (ml *MultiListener) OnDrop(element interface{}) {
	for _, l := range ml.ls {
		l.OnDrop(element)
	}
}

func (ml *MultiListener) OnOut(element interface{}) {
	for _, l := range ml.ls {
		l.OnOut(element)
	}
}

// StatListener make summation
type StatListener struct {
	incnt      uint64
	outcnt     uint64
	dropcnt    uint64
	consecdrop uint64
}

func NewStatLister() *StatListener {
	return &StatListener{}
}

func (l *StatListener) OnIn(element interface{}) {
	l.incnt++
}

func (l *StatListener) OnDrop(element interface{}) {
	l.dropcnt++
	l.consecdrop++
}

func (l *StatListener) OnOut(element interface{}) {
	l.outcnt++
	l.consecdrop = 0
}
