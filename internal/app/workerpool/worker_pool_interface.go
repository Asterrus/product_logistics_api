package workerpool

type FakeWorkerPoolInterface interface {
	Submit(func())
	GetCallCount() uint64
}
