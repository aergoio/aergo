package p2putil

// PressableQueue non-threadsafe fixed size queue, implemented like circular queue
type PressableQueue struct {
	arr    []interface{}
	cap    int
	size   int
	offset int
}

// NewPressableQueue create a new queue
func NewPressableQueue(capacity int) *PressableQueue {
	return &PressableQueue{cap: capacity, arr: make([]interface{}, capacity)}
}

// Empty returns true if queue has no element, or false if not
func (q *PressableQueue) Empty() bool {
	return q.size == 0
}

// Full returns true if queue has maximum number of elements, or false if not
func (q *PressableQueue) Full() bool {
	return q.size == q.cap
}

// Size return the number of element queue has
func (q *PressableQueue) Size() int {
	return q.size
}

// Offer is adding element to queue, it returns true if add success, or false if queue if add fail.
func (q *PressableQueue) Offer(e interface{}) bool {
	if q.size < q.cap {
		idx := (q.offset + q.size) % q.cap
		q.arr[idx] = e
		q.size++
		return true
	}
	return false
}

// Press is adding element to queue and return nil fi queue is not full, or drop first element and return dropped element if queue is full.
func (q *PressableQueue) Press(e interface{}) interface{} {
	if q.size < q.cap {
		idx := (q.offset + q.size) % q.cap
		q.arr[idx] = e
		q.size++
		return nil
	}
	idx := (q.offset + q.size) % q.cap
	toDrop := q.arr[idx]
	q.arr[idx] = e
	q.offset = (idx + 1) % q.cap
	return toDrop
}

// Peek return first element but not delete in queue. It returns nil if queue is empty
func (q *PressableQueue) Peek() interface{} {
	if q.size == 0 {
		return nil
	}
	return q.arr[q.offset]
}

// Poll return first element and remove it in queue. It returns nil if queue is empty
func (q *PressableQueue) Poll() interface{} {
	if q.size == 0 {
		return nil
	}
	e := q.arr[q.offset]
	q.arr[q.offset] = nil
	q.offset = (q.offset + 1) % q.cap
	q.size--
	return e
}
