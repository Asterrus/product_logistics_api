package consumer

import (
	"context"
	"log"
	"product_logistics_api/internal/app/queue"
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
	processedEvents <-chan model.ProductEventProcessed

	repo ports.EventRepo

	batchSize uint64
	timeout   time.Duration

	wg         *sync.WaitGroup
	eventQueue queue.EventQueue
}

func NewDbConsumer(
	n uint64,
	batchSize uint64,
	consumeTimeout time.Duration,
	repo ports.EventRepo,
	processedEvents <-chan model.ProductEventProcessed,
	eventQueue queue.EventQueue,
) Consumer {
	wg := &sync.WaitGroup{}
	return &consumer{
		n:               n,
		batchSize:       batchSize,
		timeout:         consumeTimeout,
		repo:            repo,
		processedEvents: processedEvents,
		wg:              wg,
		eventQueue:      eventQueue,
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
						default:
							c.eventQueue.PushEvent(&event)
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
