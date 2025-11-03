package producer

import (
	"context"
	"log"
	"product_logistics_api/internal/app/queue"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Producer interface {
	Start(context.Context)
	Close()
}

type producer struct {
	n uint64

	sender          ports.EventSender
	processedEvents chan<- model.ProductEventProcessed

	workerPool ports.TaskSubmitter

	wg         *sync.WaitGroup
	timeout    time.Duration
	eventQueue queue.EventQueue
}

func NewKafkaProducer(
	n uint64,
	sender ports.EventSender,
	timeout time.Duration,
	processedEvents chan<- model.ProductEventProcessed,
	workerPool ports.TaskSubmitter,
	eventQueue queue.EventQueue,
) Producer {

	wg := &sync.WaitGroup{}

	return &producer{
		n:               n,
		sender:          sender,
		workerPool:      workerPool,
		wg:              wg,
		processedEvents: processedEvents,
		eventQueue:      eventQueue,
		timeout:         timeout,
	}
}
func (p *producer) Start(ctx context.Context) {
	log.Println("producer START")
	for i := uint64(0); i < p.n; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				if !p.eventQueue.WaitForEvent(ctx) {
					log.Println("Producer stopped by context:", ctx.Err())
					return
				}
				for {
					event := p.eventQueue.PopEvent()
					if event == nil {
						break
					}
					log.Printf("Producer. Event received %v\n", event)
					var eventResult model.EventProcessedResult
					if err := p.sender.Send(event); err != nil {
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
