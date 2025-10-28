package producer

import (
	"context"
	"product_logistics_api/internal/app/sender"
	"product_logistics_api/internal/app/workerpool"
	"product_logistics_api/internal/model"
	"testing"
	"time"
)

func createProduct(ID uint64) *model.Product {
	return &model.Product{
		ID: ID,
	}
}

func createEvent(
	EventID uint64,
	Type model.EventType,
	Status model.EventStatus,
	Product *model.Product,
) *model.ProductEvent {
	return &model.ProductEvent{
		ID:     EventID,
		Type:   Type,
		Status: Status,
		Entity: Product,
	}
}

func TestStartClose(t *testing.T) {
	t.Parallel()
	producerWorkersCount := uint64(1)
	sender := sender.NewProductEventSender()
	eventsChannel := make(chan model.ProductEvent, 1)
	workerPool := workerpool.NewFakeWorkerPool()
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	producer := NewKafkaProducer(producerWorkersCount, sender, eventsChannel, processedEventsChannel, workerPool)

	ctx, cancel := context.WithCancel(context.Background())
	producer.Start(ctx)
	cancel()
	producer.Close()
}

func TestCorrectWork(t *testing.T) {
	t.Parallel()
	producerWorkersCount := uint64(1)
	sender := sender.NewProductEventSender()
	eventsChannel := make(chan model.ProductEvent, 1)
	workerPool := workerpool.NewFakeWorkerPool()
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	producer := NewKafkaProducer(producerWorkersCount, sender, eventsChannel, processedEventsChannel, workerPool)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	eventsChannel <- *new_event

	ctx, cancel := context.WithCancel(context.Background())
	producer.Start(ctx)
	cancel()
	producer.Close()
	time.Sleep(time.Millisecond * 100)

	if expectedEventsCount := 0; len(eventsChannel) != expectedEventsCount {
		t.Errorf("Events count in events channel. Expected: %v, Found: %v",
			expectedEventsCount, len(eventsChannel))
	}

	if expectedSendCount := uint64(1); sender.SendsCount() != expectedSendCount {
		t.Errorf("Producer sender.Send method call count. Expected: %v, Found: %v",
			expectedSendCount, sender.SendsCount())
	}

	if expectedCallCount := uint64(1); workerPool.GetCallCount() != expectedCallCount {
		t.Errorf("Submit call count. Expected: %v, Found: %v",
			expectedCallCount, workerPool.GetCallCount())
	}

	if expectedProcessedEventsCount := 1; len(processedEventsChannel) != expectedProcessedEventsCount {
		t.Errorf("Processed events count. Expected: %v, Found: %v",
			expectedProcessedEventsCount, len(processedEventsChannel))
	}

}
