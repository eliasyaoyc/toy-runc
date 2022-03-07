package queue

import (
	"sync"
	"time"
)

type Queue interface {
	Add(elem interface{})
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	Len() int
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

func New() *WorkQueue {
	return newWorkQueue(defaultUnfinishedWorkUpdatePeriod)
}

type t interface{}
type empty struct{}
type set map[t]empty

const defaultUnfinishedWorkUpdatePeriod = 500 * time.Millisecond

type WorkQueue struct {
	queue                      []t
	dirty                      set
	processing                 set
	cond                       *sync.Cond
	shuttingDown               bool
	drain                      bool
	unfinishedWorkUpdatePeriod time.Duration
	clock                      time.Ticker
}

func newWorkQueue(updatePeriod time.Duration) *WorkQueue {
	wq := &WorkQueue{
		dirty:                      set{},
		processing:                 set{},
		cond:                       sync.NewCond(&sync.Mutex{}),
		unfinishedWorkUpdatePeriod: updatePeriod,
		clock:                      time.Ticker{},
	}
	return wq
}

func (s set) has(item t) bool {
	_, exist := s[item]
	return exist
}

func (s set) inset(item t) {
	s[item] = empty{}
}

func (s set) delete(item t) {
	delete(s, item)
}

func (s set) len() int {
	return len(s)
}

// Add marks item as needing processing.
func (q *WorkQueue) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}

	if q.dirty.has(item) {
		return
	}

	q.dirty.inset(item)

	if q.processing.has(item) {
		return
	}

	q.queue = append(q.queue, item)
	q.cond.Signal()
}

func (q *WorkQueue) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		return nil, true
	}
	item = q.queue[0]
	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	q.queue[0] = nil
	q.queue = q.queue[1:]

	q.processing.inset(item)
	q.dirty.delete(item)
	return item, false
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing
func (q *WorkQueue) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.delete(item)
	if q.dirty.has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

// Len returns the current queue length, for informational purposes only. You
// shouldn't e.g. gate a call to Add() or Get() on Len() being a particular
// value, that can't be synchronized properly.
func (q *WorkQueue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

// ShutDown will cause q to ignore all new items added to it and
// immediately instruct the worker goroutines to exit.
func (q *WorkQueue) ShutDown() {
	q.setDrain(false)
	q.shutdown()
}

// ShutDownWithDrain will cause q to ignore all new items added to it. As soon
// as the worker goroutines have "drained", i.e: finished processing and called
// Done on all existing items in the queue; they will be instructed to exit and
// ShutDownWithDrain will return. Hence: a strict requirement for using this is;
// your workers must ensure that Done is called on all items in the queue once
// the shut down jas been initialized, if that is not the case: this will block
// indefinitely. It is, however, safe to call ShutDown after having called
// ShutDownWithDrain, as to force the queue shut down to terminate immediately
// without waiting for the drainage.
func (q *WorkQueue) ShutDownWithDrain() {
	q.setDrain(true)
	q.shutdown()
	for q.isProcessing() && q.shouldDrain() {
		q.waitForProcessing()
	}
}

func (q *WorkQueue) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.shuttingDown
}

// isProcessing indicates if there are still items on the work queue being
// processed. It's used to drain the work queue on an eventual shutdown.
func (q *WorkQueue) isProcessing() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.processing) != 0
}

// waitForProcessing waits for the worker goroutines to finish processing items
// and call Done on them.
func (q *WorkQueue) waitForProcessing() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	// Ensure that we do not wait on a queue which is already empty, as that
	// could result in waiting for Done to be called on items in an empty queue
	// which has already been shut down, which will result in waiting indefinitely.
	if len(q.processing) == 0 {
		return
	}
	q.cond.Wait()
}

func (q *WorkQueue) setDrain(shouldDrain bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.drain = shouldDrain
}

func (q *WorkQueue) shouldDrain() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.drain
}

func (q *WorkQueue) shutdown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}
