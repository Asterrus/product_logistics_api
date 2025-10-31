package retranslator

import (
	"context"
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/app/sender"
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
func TestStart(t *testing.T) {
	repo := repo.NewInMemoryProductEventRepo()
	sender := sender.NewProductEventSender()

	cfg := Config{
		EventsChannelSize:          512,
		ProcessedEventsChannelSize: 512,
		ConsumerCount:              2,
		ConsumeSize:                10,
		ConsumeTimeout:             10 * time.Second,
		ProducerCount:              2,
		WorkerCount:                2,
		DbUpdatersTimeout:          10 * time.Second,
		DbUpdatersTimeoutBatchSize: 5,
		DbUpdatersCount:            2,
		Repo:                       repo,
		Sender:                     sender,
	}

	ctx, cancel := context.WithCancel(context.Background())
	retranslator := NewRetranslator(cfg)
	retranslator.Start(ctx)
	cancel()
	retranslator.Close()
}

func TestCorrectWork(t *testing.T) {
	repo := repo.NewInMemoryProductEventRepo()
	sender := sender.NewProductEventSender()

	cfg := Config{
		EventsChannelSize:          1,
		ProcessedEventsChannelSize: 1,
		ConsumerCount:              1,
		ConsumeSize:                1,
		ConsumeTimeout:             time.Millisecond * 100,
		ProducerCount:              1,
		WorkerCount:                1,
		DbUpdatersTimeout:          time.Millisecond * 100,
		DbUpdatersTimeoutBatchSize: 1,
		DbUpdatersCount:            2,
		Repo:                       repo,
		Sender:                     sender,
	}
	ctx, cancel := context.WithCancel(context.Background())
	retranslator := NewRetranslator(cfg)
	retranslator.Start(ctx)

	defer retranslator.Close()
	defer cancel()
	time.Sleep(time.Millisecond * 25)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	time.Sleep(time.Millisecond * 150)

	if expectedEventsCount := uint64(0); repo.Count() != expectedEventsCount {
		t.Errorf("Expected event to be deleted after processing . Expected: %v, Found: %v",
			expectedEventsCount, repo.Count())
	}
}
