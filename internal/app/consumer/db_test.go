package consumer

import (
	"context"
	"product_logistics_api/internal/app/queue"
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestConsumerWork(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	consumersCount := uint64(1)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)

	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)
	cancel()
	consumer.Close()
}

func TestOneConsumerLocksCount(t *testing.T) {
	t.Parallel()

	repo := repo.NewInMemoryProductEventRepo()
	consumersCount := uint64(1)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)

	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	time.Sleep(120 * time.Millisecond)
	cancel()
	consumer.Close()
	expectedLockCalls := uint64(2)
	if repo.CountOfLocksCall() != expectedLockCalls {
		t.Errorf("Lock calls count. Expected: %d, Found: %d", expectedLockCalls, repo.CountOfLocksCall())

	}
}

func TestMultipleConsumersLocksCount(t *testing.T) {
	t.Parallel()

	repo := repo.NewInMemoryProductEventRepo()
	consumersCount := uint64(5)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)
	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	time.Sleep(60 * time.Millisecond)

	cancel()
	consumer.Close()
	expectedLockCalls := uint64(5)
	if repo.CountOfLocksCall() != expectedLockCalls {
		t.Errorf("Lock calls count. Expected: %d, Found: %d", expectedLockCalls, repo.CountOfLocksCall())

	}
}

func TestWriteInEventsQueue(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	consumersCount := uint64(1)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 10
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	defer consumer.Close()
	defer cancel()
	time.Sleep(60 * time.Millisecond)
	if queue.Len() != 1 {
		t.Errorf("Expected item in queue")
	}
}

func TestLocksStoppedAfterConsumerStopped(t *testing.T) {
	t.Parallel()

	repo := repo.NewInMemoryProductEventRepo()
	consumersCount := uint64(1)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)

	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	time.Sleep(60 * time.Millisecond)
	cancel()
	consumer.Close()
	time.Sleep(60 * time.Millisecond)

	expectedLockCalls := uint64(1)
	if repo.CountOfLocksCall() != expectedLockCalls {
		t.Errorf("Lock calls count. Expected: %d, Found: %d", expectedLockCalls, repo.CountOfLocksCall())

	}
}

func TestMultipleConsumersSendAllEvents(t *testing.T) {
	t.Parallel()

	repo := repo.NewInMemoryProductEventRepo()
	for i := range 5 {
		repo.Add(*createEvent(uint64(i), model.Created, model.Deferred, createProduct(uint64(i))))
	}
	consumersCount := uint64(5)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)
	ctx, cancel := context.WithCancel(context.Background())
	consumer.Start(ctx)

	defer consumer.Close()
	defer cancel()
	received := map[uint64]bool{}
	time.Sleep(60 * time.Millisecond)

	for i := 0; i < 5; i++ {
		event := queue.PopEvent()
		received[event.ID] = true
	}
	if len(received) != 5 {
		t.Fatalf("Events count. Expected: 5, Found: %d", len(received))
	}

}

func TestReturneProcessedEventInEventsQueue(t *testing.T) {
	t.Parallel()

	repo := repo.NewInMemoryProductEventRepo()
	repo.Add(*createEvent(uint64(1), model.Created, model.Deferred, createProduct(uint64(1))))
	consumersCount := uint64(1)
	batchSize := uint64(1)
	consumeTimeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(5)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)

	ctx, cancel := context.WithCancel(context.Background())
	processedEventsChannel <- model.ProductEventProcessed{ID: uuid.New(), EventID: uint64(1), Result: model.Returned}
	queue := queue.NewProductQueue(nil)
	consumer := NewDbConsumer(consumersCount, batchSize, consumeTimeout, repo, processedEventsChannel, queue)
	consumer.Start(ctx)
	defer consumer.Close()
	defer cancel()
	time.Sleep(time.Millisecond * 60)

	if queue.Len() != 1 {
		time.Sleep(time.Second * 1)
		t.Fatalf("Events count. Expected: 1, Found: %d", queue.Len())
	}

	event := queue.PopEvent()
	expectedStatus := model.InProgress
	if event.Status != expectedStatus {
		t.Fatalf("Events status. Expected: %v, Found: %v", expectedStatus, event.Status)

	}

}
