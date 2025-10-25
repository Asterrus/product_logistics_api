package workerpool

import (
	"sync"
	"sync/atomic"
)

// Pool — воркер-пул
type Pool struct {
	closed bool
	mu     sync.Mutex

	tasks chan func()

	maxWorkers int

	wgWorkers sync.WaitGroup // ждет завершения воркеров
	wgTasks   sync.WaitGroup // считает принятые задачи

	// атомарный флаг для быстрой проверки
	closedInt int32
}

// New создаёт пул с maxWorkers и очередью размера queueSize.
// queueSize = 0 — значит без буфера (подача будет блокирующей, если все воркеры заняты).
func New(maxWorkers int, queueSize int) *Pool {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	if queueSize < 0 {
		queueSize = 0
	}
	p := &Pool{
		tasks:      make(chan func(), queueSize),
		maxWorkers: maxWorkers,
	}
	p.startWorkers(uint64(p.maxWorkers))
	return p
}

func (p *Pool) startWorkers(_ uint64) {
	for i := 0; i < p.maxWorkers; i++ {
		p.wgWorkers.Add(1)
		go func() {
			defer p.wgWorkers.Done()
			for task := range p.tasks {
				// Безопасно: wgTasks гарантирует, что для каждой принятой задачи был вызван Add
				func() {
					defer func() {
						// защита: если task вызовет паник — всё равно считать задачу выполненной
						if r := recover(); r != nil {
							// можно логировать, если нужно
						}
					}()
					task()
				}()
				p.wgTasks.Done()
			}
		}()
	}
}

// Submit — блокирующая подача задачи.
// Возвращает ErrPoolStopped если пул уже остановлен.
func (p *Pool) Submit(task func()) error {
	// quick check
	if atomic.LoadInt32(&p.closedInt) == 1 {
		return ErrPoolStopped
	}

	// защита от гонки закрытия: берем муьтекс для синхронной проверки+Add
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolStopped
	}
	p.wgTasks.Add(1)
	p.mu.Unlock()

	// безопасно отправляем (send может паникнуть, если между Add и send кто-то закрыл канал)
	if err := p.sendSafely(task, false); err != nil {
		// вернуть счетчик задач, т.к. задача не была принята
		p.wgTasks.Done()
		return err
	}
	return nil
}

// TrySubmit — неблокирующая подача; если очередь полна — возвращает ErrQueueFull.
// Возвращает ErrPoolStopped если пул уже остановлен.
func (p *Pool) TrySubmit(task func()) error {
	if atomic.LoadInt32(&p.closedInt) == 1 {
		return ErrPoolStopped
	}

	// блокируем на время проверки + Add, чтобы избежать race с Stop
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolStopped
	}
	p.wgTasks.Add(1)
	p.mu.Unlock()

	// делаем неблокирующий send, но в безопасной оболочке
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = ErrPoolStopped
			}
		}()
		select {
		case p.tasks <- task:
			return nil
		default:
			return ErrQueueFull
		}
	}()

	if err != nil {
		p.wgTasks.Done()
	}
	return err
}

// sendSafely делает блокирующую отправку в p.tasks и ловит паники (закрытие канала)
func (p *Pool) sendSafely(task func(), _unused bool) error {
	defer func() {
		// recover в вызывавшей функции не нужен здесь, т.к. мы обрабатываем внутри
	}()
	// ловим панику, которая может возникнуть при отправке в закрытый канал
	defer func() {
		if r := recover(); r != nil {
			// nothing here; we handle via returned error below by using named return in wrapper
		}
	}()
	// Но recover не может установить ошибку здесь напрямую, поэтому используем вложенную функцию:
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = ErrPoolStopped
			}
		}()
		p.tasks <- task
		return nil
	}()
	return err
}

// Stop — инициирует остановку пула: пул больше не принимает задачи.
// Текущие принятые задачи будут выполнены.
// Stop не ждёт окончания выполнения — для ожидания используйте StopWait.
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	atomic.StoreInt32(&p.closedInt, 1)
	// закрываем канал задач — после этого воркеры отработают имеющиеся задачи и завершат range
	close(p.tasks)
	p.mu.Unlock()
}

// StopWait — аккуратно останавливает пул и ждёт, пока все принятые задачи и воркеры завершат работу.
func (p *Pool) StopWait() error {
	p.Stop()
	// дождаться, пока все воркеры закроют свои goroutine
	p.wgWorkers.Wait()
	// дождаться, пока все принятые задачи будут помечены как Done
	p.wgTasks.Wait()
	return nil
}
