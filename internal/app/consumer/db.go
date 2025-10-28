package consumer

import (
	"context"
	"log"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"sync"
	"time"
)

type Consumer interface {
	Start(context.Context)
	Close()
}

type consumer struct {
	n               uint64
	events          chan<- model.ProductEvent
	processedEvents <-chan model.ProductEventProcessed

	repo ports.EventRepo

	batchSize uint64
	timeout   time.Duration

	wg *sync.WaitGroup
}

func NewDbConsumer(
	n uint64,
	batchSize uint64,
	consumeTimeout time.Duration,
	repo ports.EventRepo,
	events chan<- model.ProductEvent,
	processedEvents <-chan model.ProductEventProcessed,
) Consumer {
	wg := &sync.WaitGroup{}
	return &consumer{
		n:               n,
		batchSize:       batchSize,
		timeout:         consumeTimeout,
		repo:            repo,
		events:          events,
		processedEvents: processedEvents,
		wg:              wg,
	}
}

func (c *consumer) Start(ctx context.Context) {
	log.Println("NewDbConsumer START")
	for i := uint64(0); i < c.n; i++ {
		c.wg.Add(1)

		go func() {
			defer c.wg.Done()
			ticker := time.NewTicker(c.timeout)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					log.Println("Consumer stopped by context:", ctx.Err())
					return
				case <-ticker.C:
					log.Println("NewDbConsumer case ticker.C")
					events, err := c.repo.Lock(c.batchSize)
					log.Printf("NewDbConsumer events: %v, err: %v\n", events, err)
					if err != nil {
						log.Printf("Repo Lock error: %v", err)
						continue
					}
					for _, event := range events {
						log.Printf("NewDbConsumer Send event %v to events channel\n", event)
						select {
						case <-ctx.Done():
							return
						case c.events <- event:

						}
					}
				case e := <-c.processedEvents:
					if e.Result == model.Sent {

						if err := c.repo.Remove([]uint64{e.EventID}); err != nil {
							// Что делаем если не удалось удалить запись о событии из базы?
							log.Printf("Repo Remove error: %v", err)
						}

					} else {

						if err := c.repo.Unlock([]uint64{e.EventID}); err != nil {
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
	c.wg.Wait()
}
