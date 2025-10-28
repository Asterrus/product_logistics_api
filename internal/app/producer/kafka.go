package producer

import (
	"fmt"
	"product_logistics_api/internal/app/sender"
	"product_logistics_api/internal/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Producer interface {
	Start()
	Close()
}

type WorkerPool interface {
	Submit(task func())
}
type producer struct {
	n       uint64
	timeout time.Duration

	sender          sender.EventSender
	events          <-chan model.ProductEvent
	processedEvents chan<- model.ProductEventProcessed

	workerPool WorkerPool

	wg   *sync.WaitGroup
	done chan struct{}
}

// todo for students: add repo
func NewKafkaProducer(
	n uint64,
	sender sender.EventSender,
	events <-chan model.ProductEvent,
	processedEvents chan<- model.ProductEventProcessed,
	workerPool WorkerPool,
) Producer {

	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	return &producer{
		n:               n,
		sender:          sender,
		events:          events,
		workerPool:      workerPool,
		wg:              wg,
		done:            done,
		processedEvents: processedEvents,
	}
}
func (p *producer) Start() {
	fmt.Println("producer START")
	for i := uint64(0); i < p.n; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case event := <-p.events:
					fmt.Printf("producer. Event recieved %v\n", event)

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
				case <-p.done:
					fmt.Println("producer. p.done")

					return
				}
			}
		}()
	}
}

func (p *producer) Close() {
	fmt.Println("producer Close")
	close(p.done)
	p.wg.Wait()
}
