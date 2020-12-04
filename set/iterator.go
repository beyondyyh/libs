package set

// Iterator 定义一个基于集合的迭代器，C channel用来迭代集合的元素
type Iterator struct {
	C    <-chan interface{}
	stop chan struct{}
}

// Stop stops the Iterator, 停止迭代器并关闭channel，后续元素将不会继续输出
func (i *Iterator) Stop() {
	// 防止多次调用Stop()引起panic，(close() panics when called on already closed channel)
	// panic info: close of closed channel
	defer func() {
		recover()
	}()

	close(i.stop)

	// 丢弃未输出的元素
	for range i.C {
	}
}

// newIterator returns a new Iterator instance
func newIterator() (*Iterator, chan<- interface{}, <-chan struct{}) {
	itemChan := make(chan interface{})
	stopChan := make(chan struct{})
	return &Iterator{
		C:    itemChan,
		stop: stopChan,
	}, itemChan, stopChan
}
