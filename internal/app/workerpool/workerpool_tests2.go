package workerpool

// import (
// 	"errors"
// 	"sync"
// 	"sync/atomic"
// 	"testing"
// 	"time"
// )

// type Factory struct {
// 	name    string
// 	factory *PoolFactory
// }

// // PoolFactory — создаёт Pool для теста.
// // Если реализация не возвращает ошибку, factory должна вернуть (Pool, nil).
// type PoolFactory func(workers, queueSize int) (Pool, error)

// // Список фабрик — добавляй сюда реализации, которые хочешь тестировать.
// func factories() []Factory {
// 	return []Factory{
// 		{"advanced", func(w, q int) (Pool, error) { return New(w, q), nil }},
// 		{"simple", func(w, q int) (Pool, error) { return NewSimple(w, q) }},
// 	}
// }

// // runForAll — запускает под-тест для каждой фабрики
// func runForAll(t *testing.T, name string, fn func(t *testing.T, factory PoolFactory)) {
// 	for _, f := range factories() {
// 		t.Run(f.name+"/"+name, func(t *testing.T) {
// 			fn(t, f.factory)
// 		})
// 	}
// }

// func TestBasicProcessing(t *testing.T) {
// 	runForAll(t, "BasicProcessing", func(t *testing.T, factory PoolFactory) {
// 		p, err := factory(3, 5)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		var counter int32
// 		tasks := 10
// 		for i := 0; i < tasks; i++ {
// 			if err := p.Submit(func() {
// 				atomic.AddInt32(&counter, 1)
// 				time.Sleep(5 * time.Millisecond)
// 			}); err != nil {
// 				t.Fatal("Submit failed:", err)
// 			}
// 		}
// 		p.StopWait()
// 		if got := atomic.LoadInt32(&counter); got != int32(tasks) {
// 			t.Fatalf("expected %d tasks executed, got %d", tasks, got)
// 		}
// 	})
// }

// func TestPanicDoesNotKillWorkers(t *testing.T) {
// 	runForAll(t, "PanicDoesNotKillWorkers", func(t *testing.T, factory PoolFactory) {
// 		p, err := factory(2, 5)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		var done int32
// 		_ = p.Submit(func() { panic("boom") })
// 		_ = p.Submit(func() { atomic.StoreInt32(&done, 1) })
// 		p.StopWait()
// 		if atomic.LoadInt32(&done) != 1 {
// 			t.Fatal("task after panic did not execute")
// 		}
// 	})
// }

// func TestAddAfterStop(t *testing.T) {
// 	runForAll(t, "AddAfterStop", func(t *testing.T, factory PoolFactory) {
// 		p, err := factory(1, 2)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		p.Stop()
// 		// Submit should return ErrPoolStopped
// 		if err := p.Submit(func() {}); !errors.Is(err, ErrPoolStopped) {
// 			t.Fatalf("expected ErrPoolStopped for Submit, got %v", err)
// 		}
// 		// TrySubmit too
// 		if err := p.TrySubmit(func() {}); !errors.Is(err, ErrPoolStopped) {
// 			t.Fatalf("expected ErrPoolStopped for TrySubmit, got %v", err)
// 		}
// 	})
// }

// func TestTrySubmitQueueFull(t *testing.T) {
// 	runForAll(t, "TrySubmitQueueFull", func(t *testing.T, factory PoolFactory) {
// 		p, err := factory(2, 2)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		var executed int32
// 		// fill up queue and workers: 2 workers + 2 queue = 4 tasks accepted
// 		for i := 0; i < 4; i++ {
// 			if err := p.TrySubmit(func() { atomic.AddInt32(&executed, 1); time.Sleep(20 * time.Millisecond) }); err != nil {
// 				t.Fatal("TrySubmit failed unexpectedly:", err)
// 			}
// 		}
// 		// next TrySubmit must return ErrQueueFull (non-blocking)
// 		if err := p.TrySubmit(func() {}); !errors.Is(err, ErrQueueFull) {
// 			t.Fatalf("expected ErrQueueFull, got %v", err)
// 		}
// 		p.StopWait()
// 		if atomic.LoadInt32(&executed) != 4 {
// 			t.Fatalf("expected 4 executed tasks, got %d", executed)
// 		}
// 	})
// }

// func TestConcurrentLimit(t *testing.T) {
// 	runForAll(t, "ConcurrentLimit", func(t *testing.T, factory PoolFactory) {
// 		workers := 3
// 		p, err := factory(workers, 10)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		var maxConcurrent int32
// 		var current int32
// 		total := 30
// 		var wg sync.WaitGroup
// 		wg.Add(total)
// 		for i := 0; i < total; i++ {
// 			if err := p.Submit(func() {
// 				c := atomic.AddInt32(&current, 1)
// 				for {
// 					old := atomic.LoadInt32(&maxConcurrent)
// 					if c <= old {
// 						break
// 					}
// 					if atomic.CompareAndSwapInt32(&maxConcurrent, old, c) {
// 						break
// 					}
// 				}
// 				time.Sleep(5 * time.Millisecond)
// 				atomic.AddInt32(&current, -1)
// 				wg.Done()
// 			}); err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 		wg.Wait()
// 		p.StopWait()
// 		if atomic.LoadInt32(&maxConcurrent) > int32(workers) {
// 			t.Fatalf("max concurrent %d > workers %d", maxConcurrent, workers)
// 		}
// 	})
// }
