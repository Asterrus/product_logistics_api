package consumer

import (
	"fmt"
	"log"
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/model"
	"sync"
	"time"
)

type Consumer interface {
	Start()
	Close()
}

type consumer struct {
	n               uint64
	events          chan<- model.ProductEvent
	processedEvents <-chan model.ProductEventProcessed

	repo repo.EventRepo

	batchSize uint64
	timeout   time.Duration

	done chan struct{}
	wg   *sync.WaitGroup
}

func NewDbConsumer(
	n uint64,
	batchSize uint64,
	consumeTimeout time.Duration,
	repo repo.EventRepo,
	events chan<- model.ProductEvent,
	processedEvents <-chan model.ProductEventProcessed,
) Consumer {
	wg := &sync.WaitGroup{}
	done := make(chan struct{})
	return &consumer{
		n:               n,
		batchSize:       batchSize,
		timeout:         consumeTimeout,
		repo:            repo,
		events:          events,
		processedEvents: processedEvents,
		done:            done,
		wg:              wg,
	}
}

func (c *consumer) Start() {
	fmt.Println("NewDbConsumer START")
	for i := uint64(0); i < c.n; i++ {
		c.wg.Add(1)

		go func() {
			defer c.wg.Done()
			ticker := time.NewTicker(c.timeout)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					fmt.Println("NewDbConsumer case ticker.C")
					events, err := c.repo.Lock(c.batchSize)
					fmt.Printf("NewDbConsumer events: %v, err: %v\n", events, err)
					if err != nil {
						log.Printf("Repo Lock error: %v", err)
						continue
					}
					for _, event := range events {
						fmt.Printf("NewDbConsumer Send event %v to events channel\n", event)
						c.events <- event
					}
				case <-c.done:
					fmt.Println("NewDbConsumer case c.done.")
					return
				case e := <-c.processedEvents:
					if e.Result == model.Sent {
						err := c.repo.Remove([]uint64{e.EventID})
						if err != nil {
							// Что делаем если не удалось удалить запись о событии из базы?
							log.Printf("Repo Remove error: %v", err)
						}

					} else {
						err := c.repo.Unlock([]uint64{e.EventID})
						if err != nil {
							// Что делаем если не вышло вернуть записи статус "К обработке"?
							log.Printf("Repo Unlock error: %v", err)
						}
					}

				}
			}
		}()
	}
}

func (c *consumer) Close() {
	fmt.Println("NewDbConsumer Close")
	close(c.done)
	c.wg.Wait()
}
