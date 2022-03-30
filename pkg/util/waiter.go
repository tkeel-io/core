package util

type Waiter interface {
	Add(int)
	Done()
	Wait()
}

type waitGrop struct {
}

func NewWaiter() Waiter {
	return &waitGrop{}
}

func (wg *waitGrop) Add(int) {}
func (wg *waitGrop) Done()   {}
func (wg *waitGrop) Wait()   {}
