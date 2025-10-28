package producer

import (
	"context"
	"log"
	"product_logistics_api/internal/app/sender"
	"product_logistics_api/internal/model"
	"sync"

	"github.com/google/uuid"
)

type Producer interface {
	Start(context.Context)
	Close()
}

type WorkerPool interface {
	Submit(task func())
}
type producer struct {
	n uint64

	sender          sender.EventSender
	events          <-chan model.ProductEvent
	processedEvents chan<- model.ProductEventProcessed

	workerPool WorkerPool

	wg *sync.WaitGroup
}

func NewKafkaProducer(
	n uint64,
	sender sender.EventSender,
	events <-chan model.ProductEvent,
	processedEvents chan<- model.ProductEventProcessed,
	workerPool WorkerPool,
) Producer {

	wg := &sync.WaitGroup{}

	return &producer{
		n:               n,
		sender:          sender,
		events:          events,
		workerPool:      workerPool,
		wg:              wg,
		processedEvents: processedEvents,
	}
}
func (p *producer) Start(ctx context.Context) {
	log.Println("producer START")
	for i := uint64(0); i < p.n; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Println("Producer stopped by context:", ctx.Err())
					return
				case event := <-p.events:
					log.Printf("Producer. Event recieved %v\n", event)

					if err := p.sender.Send(&event); err != nil {
						p.workerPool.Submit(func() {
							p.processedEvents <- model.ProductEventProcessed{
								ID:      uuid.New(),
								EventID: event.ID,
								Result:  model.Returned,
							}
						})
					} else {
						p.workerPool.Submit(func() {
							p.processedEvents <- model.ProductEventProcessed{
								ID:      uuid.New(),
								EventID: event.ID,
								Result:  model.Sent,
							}
						})
					}
				}
			}
		}()
	}
}

func (p *producer) Close() {
	p.wg.Wait()
}
