package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func consumer(q *Queue, t *testing.T) {
	for {
		x := q.Pop()
		t.Logf("Read from queue: %v", x)
	}
}

func producer(q *Queue, t *testing.T) {
	assert := assert.New(t)
	for i := 0; i < 10; i = i + 1 {
		err := q.Push(i)
		assert.Nil(err, "q.Push() should return nil")
		t.Logf("Write to queue: %d", i)
	}
}

// run: go test -v -run TestSendQueue
func TestSendQueue(t *testing.T) {
	var q Queue
	q.Init()

	go consumer(&q, t)
	go producer(&q, t)

	time.Sleep(2 * time.Second)
}

// run: go test -v -run TestQueueIsFull
func TestQueueIsFull(t *testing.T) {
	var q Queue
	q.Init()
	q.SetMaxLen(3)

	assert := assert.New(t)
	for i := 0; i < 10; i = i + 1 {
		err := q.Push(i)
		if i < 3 {
			assert.Nil(err, "q.Push() should return nil")
		} else {
			// assert.NotNil(err, "q.Push() should return error")
			assert.EqualError(err, "Queue is full")
		}
	}
}
