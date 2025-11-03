package producer

import (
	"context"
	"product_logistics_api/internal/app/queue"
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
	workerPool := workerpool.NewFakeWorkerPool()
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	timeout := time.Millisecond * 50
	producer := NewKafkaProducer(producerWorkersCount, sender, timeout, processedEventsChannel, workerPool, queue)

	ctx, cancel := context.WithCancel(context.Background())
	producer.Start(ctx)
	cancel()
	producer.Close()
}

func TestCorrectWork(t *testing.T) {
	t.Parallel()
	producerWorkersCount := uint64(1)
	sender := sender.NewProductEventSender()
	workerPool := workerpool.NewFakeWorkerPool()
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	timeout := time.Millisecond * 50
	producer := NewKafkaProducer(producerWorkersCount, sender, timeout, processedEventsChannel, workerPool, queue)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	queue.PushEvent(new_event)

	ctx, cancel := context.WithCancel(context.Background())
	producer.Start(ctx)

	defer producer.Close()
	defer cancel()

	time.Sleep(time.Millisecond * 100)

	if expectedEventsCount := 0; queue.Len() != expectedEventsCount {
		t.Errorf("Events count in events channel. Expected: %v, Found: %v",
			expectedEventsCount, queue.Len())
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
