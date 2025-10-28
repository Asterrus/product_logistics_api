package workerpool

import (
	"errors"
	"log"
	"sync"
)

type WorkerPool struct {
	closed bool
	mu     sync.Mutex

	tasks chan func()

	doneCh chan struct{}
}

// Ошибки
var (
	ErrPoolStopped = errors.New("workerpool: pool stopped")
	ErrQueueFull   = errors.New("workerpool: queue full")
)

func NewWorkerPool(workersNum uint64, tasksBufferSize uint64) (*WorkerPool, error) {
	if workersNum <= 0 {
		return nil, errors.New("incorrect workers num")
	}
	wp := &WorkerPool{
		tasks:  make(chan func(), tasksBufferSize),
		doneCh: make(chan struct{}),
	}
	workersInitWG := sync.WaitGroup{}
	workersInitWG.Add(int(workersNum))
	go wp.startWorkers(workersNum, &workersInitWG)
	workersInitWG.Wait()

	return wp, nil
}

func (wp *WorkerPool) startWorkers(workersNum uint64, workersInitWG *sync.WaitGroup) {
	wg := sync.WaitGroup{}

	for i := 0; i < int(workersNum); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workersInitWG.Done()
			for task := range wp.tasks {

				func(i int) {
					log.Printf("WORKER %v GET TASK", i)
					defer func() {
						// защита: если task вызовет паник — всё равно считать задачу выполненной
						if r := recover(); r != nil {
							log.Printf("Panic during task execution!: %v", r)
							// можно логировать, если нужно
						}
					}()
					task()
				}(i)
			}
		}()
	}
	wg.Wait()
	close(wp.doneCh)
}

func (wp *WorkerPool) Submit(task func()) error {
	if task == nil {
		return errors.New("Task is incorrect")
	}

	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.closed {
		return ErrPoolStopped
	}
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = ErrPoolStopped
			}
		}()
		wp.tasks <- task
		return nil
	}()

	return err
}

func (wp *WorkerPool) TrySubmit(task func()) error {
	if task == nil {
		return errors.New("Task is incorrect")
	}

	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.closed {
		return ErrPoolStopped
	}
	select {
	case wp.tasks <- task:
		return nil
	default:
		return ErrQueueFull
	}
}
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.closed {
		return ErrPoolStopped

	}
	wp.closed = true
	close(wp.tasks)
	return nil
}
func (wp *WorkerPool) StopWait() error {
	wp.Stop()
	<-wp.doneCh
	return nil
}

// Проверить 	case wp.tasks <- task: Если канал вдруг закрыт
