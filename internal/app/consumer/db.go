package consumer

import (
	"fmt"
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
	n      uint64
	events chan<- model.ProductEvent

	repo repo.EventRepo

	batchSize uint64
	timeout   time.Duration

	done chan bool
	wg   *sync.WaitGroup
}

func NewDbConsumer(
	n uint64,
	batchSize uint64,
	consumeTimeout time.Duration,
	repo repo.EventRepo,
	events chan<- model.ProductEvent,
) Consumer {
	wg := &sync.WaitGroup{}
	done := make(chan bool)
	return &consumer{
		n:         n,
		batchSize: batchSize,
		timeout:   consumeTimeout,
		repo:      repo,
		events:    events,
		done:      done,
		wg:        wg,
	}
}

func (c *consumer) Start() {
	fmt.Println("NewDbConsumer START")
	for i := uint64(0); i < c.n; i++ {
		c.wg.Add(1)

		go func() {
			defer c.wg.Done()
			ticker := time.NewTicker(c.timeout)
			for {
				select {
				case <-ticker.C:
					fmt.Println("NewDbConsumer case ticker.C")
					events, err := c.repo.Lock(c.batchSize)
					fmt.Printf("NewDbConsumer events: %v, err: %v\n", events, err)
					if err != nil {
						panic(err)
						// continue
					}
					for _, event := range events {
						fmt.Printf("NewDbConsumer Send event %v to events channel\n", event)
						c.events <- event
					}
				case <-c.done:
					fmt.Println("NewDbConsumer case c.done.")
					return
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
