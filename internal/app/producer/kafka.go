package producer

import (
	"context"
	"log"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"sync"

	"github.com/google/uuid"
)

type Producer interface {
	Start(context.Context)
	Close()
}

type producer struct {
	n uint64

	sender          ports.EventSender
	events          <-chan model.ProductEvent
	processedEvents chan<- model.ProductEventProcessed

	workerPool ports.TaskSubmitter

	wg *sync.WaitGroup
}

func NewKafkaProducer(
	n uint64,
	sender ports.EventSender,
	events <-chan model.ProductEvent,
	processedEvents chan<- model.ProductEventProcessed,
	workerPool ports.TaskSubmitter,
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
					log.Printf("Producer. Event received %v\n", event)
					var eventResult model.EventProcessedResult
					if err := p.sender.Send(&event); err != nil {
						log.Printf("Producer. Event send error %v\n", err)
						eventResult = model.Returned
					} else {
						eventResult = model.Sent
					}

					select {
					case <-ctx.Done():
						log.Println("Producer stopped by context:", ctx.Err())
						return
					default:
						p.workerPool.Submit(func() {
							p.processedEvents <- model.ProductEventProcessed{
								ID:      uuid.New(),
								EventID: event.ID,
								Result:  eventResult,
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
