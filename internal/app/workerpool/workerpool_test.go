package workerpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type WorkerPoolInterface interface {
	startWorkers(workersNum uint64)
	Submit(task func()) error
	StopWait() error
}

func GetMyPoolRealisation(workersNum uint64, tasksBufferSize uint64) (*WorkerPool, error) {
	return NewWorkerPool(workersNum, tasksBufferSize)
}

func TestBasicProcessing(t *testing.T) {
	p, _ := GetMyPoolRealisation(3, 5)
	// p := GetAiPoolRealisation(3, 5)
	var counter int32
	tasks := 10

	for i := 0; i < tasks; i++ {
		if err := p.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(5 * time.Millisecond)
		}); err != nil {
			t.Fatal("Submit failed:", err)
		}
	}
	p.StopWait()
	if got := atomic.LoadInt32(&counter); got != int32(tasks) {
		t.Fatalf("expected %d tasks executed, got %d", tasks, got)
	}
}

func TestPanicDoesNotKillWorkers(t *testing.T) {
	p, _ := GetMyPoolRealisation(3, 5)
	// p := GetAiPoolRealisation(3, 5)

	var done int32
	_ = p.Submit(func() { panic("boom") })
	_ = p.Submit(func() { atomic.StoreInt32(&done, 1) })
	p.StopWait()
	if atomic.LoadInt32(&done) != 1 {
		t.Fatal("task after panic did not execute")
	}

}

func TestAddAfterStop(t *testing.T) {
	p, _ := GetMyPoolRealisation(3, 5)
	// p := GetAiPoolRealisation(3, 5)
	p.Stop()
	// Submit should return ErrPoolStopped
	if err := p.Submit(func() {}); !errors.Is(err, ErrPoolStopped) {
		t.Fatalf("expected ErrPoolStopped for Submit, got %v", err)
	}
	// TrySubmit too
	if err := p.TrySubmit(func() {}); !errors.Is(err, ErrPoolStopped) {
		t.Fatalf("expected ErrPoolStopped for TrySubmit, got %v", err)
	}

}

func TestTrySubmitQueueFull(t *testing.T) {
	p, _ := GetMyPoolRealisation(2, 2)
	// p := GetAiPoolRealisation(2, 2)
	var executed int32
	// fill up queue and workers: 2 workers + 2 queue = 4 tasks accepted
	for i := 0; i < 4; i++ {
		if err := p.TrySubmit(func() { atomic.AddInt32(&executed, 1); time.Sleep(20 * time.Millisecond) }); err != nil {
			t.Fatal("TrySubmit failed unexpectedly:", err)
		}
	}
	// next TrySubmit must return ErrQueueFull (non-blocking)
	if err := p.TrySubmit(func() {}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	p.StopWait()
	if atomic.LoadInt32(&executed) != 4 {
		t.Fatalf("expected 4 executed tasks, got %d", executed)
	}

}

func TestConcurrentLimit(t *testing.T) {

	workers := 3

	p, _ := GetMyPoolRealisation(uint64(workers), 10)
	// p := GetAiPoolRealisation(workers, 10)

	var maxConcurrent int32
	var current int32
	total := 30
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		if err := p.Submit(func() {
			c := atomic.AddInt32(&current, 1)
			for {
				old := atomic.LoadInt32(&maxConcurrent)
				if c <= old {
					break
				}
				if atomic.CompareAndSwapInt32(&maxConcurrent, old, c) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&current, -1)
			wg.Done()
		}); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	p.StopWait()
	if atomic.LoadInt32(&maxConcurrent) > int32(workers) {
		t.Fatalf("max concurrent %d > workers %d", maxConcurrent, workers)
	}

}
