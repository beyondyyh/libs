package queue

import (
	"container/list"
	"errors"
	"sync"
)

// DESCRIPTION
// Usage:
//     var q queue.Queue
//     q.Init()
//     q.Push("abcd")
//     len := q.Len()
//     msg := q.Pop()
//     // type convert is required here
//     x := msg.(string)

type Queue struct {
	sync.Mutex
	cond   *sync.Cond
	list   *list.List
	maxLen int // max queue length
}

// Initialize the queue
func (q *Queue) Init() {
	q.cond = sync.NewCond(&q.Mutex)
	q.list = list.New()
	q.maxLen = -1
}

// Set queue max length
func (q *Queue) SetMaxLen(maxLen int) {
	q.Lock()
	q.maxLen = maxLen
	q.Unlock()
}

// Push an item to the queue
func (q *Queue) Push(x interface{}) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	var err error
	if q.maxLen != -1 && q.list.Len() >= q.maxLen {
		err = errors.New("Queue is full")
	} else {
		q.list.PushBack(x)
		q.cond.Signal()
	}
	return err
}

// Pop an item from the queue
func (q *Queue) Pop() interface{} {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.list.Len() == 0 {
		q.cond.Wait()
	}
	x := q.list.Front()
	q.list.Remove(x)
	return x.Value
}

// Returns the length of queue
func (q *Queue) Len() int {
	q.Lock()
	defer q.Unlock()
	return q.list.Len()
}
