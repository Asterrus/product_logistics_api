package workerpool

import "log"

type FakeWorkerPoolInterface interface {
	Submit(func())
	GetCallCount() uint64
}
type FakeWorkerPool struct {
	submitCallCount uint64
}

func NewFakeWorkerPool() FakeWorkerPoolInterface {
	return &FakeWorkerPool{submitCallCount: 0}
}

func (wp *FakeWorkerPool) Submit(task func()) {
	log.Println("FakeWorkerPool Submit called")
	wp.submitCallCount++
	task()
}

func (wp *FakeWorkerPool) GetCallCount() uint64 {
	return wp.submitCallCount
}
