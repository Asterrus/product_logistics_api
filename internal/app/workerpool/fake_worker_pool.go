package workerpool

import (
	"log"
	"product_logistics_api/internal/ports"
)

type FakeWorkerPool struct {
	submitCallCount uint64
}
type FakeWorkerPoolInterface interface {
	ports.TaskSubmitter
	GetCallCount() uint64
}

func NewFakeWorkerPool() FakeWorkerPoolInterface {
	return &FakeWorkerPool{submitCallCount: 0}
}

func (wp *FakeWorkerPool) Submit(task func()) error {
	log.Println("FakeWorkerPool Submit called")
	wp.submitCallCount++
	task()
	return nil
}

func (wp *FakeWorkerPool) GetCallCount() uint64 {
	return wp.submitCallCount
}

func (wp *FakeWorkerPool) Stop() error {
	return nil
}
func (wp *FakeWorkerPool) StopWait() error {
	return nil
}
func (wp *FakeWorkerPool) TrySubmit(task func()) error {
	return nil
}
