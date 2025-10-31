package dbupdater

import (
	"context"
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
func TestDbUpdaterWork(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	updatersCount := uint64(1)
	batchSize := uint64(1)
	timeout := time.Millisecond * 50
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	dbUpdater := NewDbUpdater(
		updatersCount,
		processedEventsChannel,
		timeout,
		batchSize,
		repo,
	)

	ctx, cancel := context.WithCancel(context.Background())
	dbUpdater.Start(ctx)
	cancel()
	dbUpdater.Close()
}

func TestEventDeletedInRepo(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	updatersCount := uint64(1)
	batchSize := uint64(1)
	timeout := time.Millisecond * 100
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	dbUpdater := NewDbUpdater(
		updatersCount,
		processedEventsChannel,
		timeout,
		batchSize,
		repo,
	)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.InProgress, prod)
	repo.Add(*new_event)

	processedEventsChannel <- model.ProductEventProcessed{
		ID:      uuid.New(),
		EventID: 1,
		Result:  model.Sent,
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbUpdater.Start(ctx)

	time.Sleep(time.Millisecond * 50)

	cancel()
	dbUpdater.Close()

	expected := 0

	if repo.Count() != uint64(expected) {
		t.Errorf("Expected: %v. Found: %v", expected, repo.Count())
	}
}

func TestEventUpdatedInRepo(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	updatersCount := uint64(1)
	batchSize := uint64(1)
	timeout := time.Millisecond * 100
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	dbUpdater := NewDbUpdater(
		updatersCount,
		processedEventsChannel,
		timeout,
		batchSize,
		repo,
	)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.InProgress, prod)
	repo.Add(*new_event)

	processedEventsChannel <- model.ProductEventProcessed{
		ID:      uuid.New(),
		EventID: 1,
		Result:  model.Returned,
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbUpdater.Start(ctx)

	time.Sleep(time.Millisecond * 50)

	cancel()
	dbUpdater.Close()

	expectedCount := 1

	if repo.Count() != uint64(expectedCount) {
		t.Errorf("Expected: %v. Found: %v", expectedCount, repo.Count())
	}
	expectedStatus := model.Deferred

	if e, _ := repo.Get(1); e.Status != expectedStatus {
		t.Errorf("Expected: %v. Found: %v", expectedStatus, e.Status)

	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	updatersCount := uint64(1)
	batchSize := uint64(2)
	timeout := time.Millisecond * 100
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	dbUpdater := NewDbUpdater(
		updatersCount,
		processedEventsChannel,
		timeout,
		batchSize,
		repo,
	)

	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.InProgress, prod)
	repo.Add(*new_event)

	processedEventsChannel <- model.ProductEventProcessed{
		ID:      uuid.New(),
		EventID: 1,
		Result:  model.Returned,
	}

	ctx, cancel := context.WithCancel(context.Background())
	dbUpdater.Start(ctx)

	time.Sleep(time.Millisecond * 150)

	cancel()
	dbUpdater.Close()

	expectedCount := 1

	if repo.Count() != uint64(expectedCount) {
		t.Errorf("Expected: %v. Found: %v", expectedCount, repo.Count())
	}
	expectedStatus := model.Deferred

	if e, _ := repo.Get(1); e.Status != expectedStatus {
		t.Errorf("Expected: %v. Found: %v", expectedStatus, e.Status)

	}
}

func TestBatch(t *testing.T) {
	t.Parallel()
	repo := repo.NewInMemoryProductEventRepo()
	updatersCount := uint64(1)
	batchSize := uint64(2)
	timeout := time.Millisecond * 100
	processedEventsChannelBuffer := uint64(1)
	processedEventsChannel := make(chan model.ProductEventProcessed, processedEventsChannelBuffer)
	dbUpdater := NewDbUpdater(
		updatersCount,
		processedEventsChannel,
		timeout,
		batchSize,
		repo,
	)

	prod := createProduct(1)
	created_event := createEvent(1, model.Created, model.InProgress, prod)
	updated_event := createEvent(2, model.Updated, model.InProgress, prod)
	repo.Add(*created_event)
	repo.Add(*updated_event)

	ctx, cancel := context.WithCancel(context.Background())
	dbUpdater.Start(ctx)

	processedEventsChannel <- model.ProductEventProcessed{
		ID:      uuid.New(),
		EventID: 1,
		Result:  model.Sent,
	}
	processedEventsChannel <- model.ProductEventProcessed{
		ID:      uuid.New(),
		EventID: 2,
		Result:  model.Returned,
	}

	time.Sleep(time.Millisecond * 150)

	cancel()
	dbUpdater.Close()

	expectedCount := 1

	if repo.Count() != uint64(expectedCount) {
		t.Errorf("Expected: %v. Found: %v", expectedCount, repo.Count())
	}
}
