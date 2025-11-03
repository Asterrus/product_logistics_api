package repo

import (
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
func TestAdd(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	expected := 1
	if repo.Count() != uint64(expected) {
		t.Errorf("Events count. Expected: %d, Found: %d", expected, repo.Count())
	}
	event, _ := repo.Get(new_event.ID)

	if event.ID != new_event.ID {
		t.Errorf("Event ID. Expected: %d, Found: %d", new_event.ID, event.ID)
	}
}

func TestLock(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	repo.Lock(1)

	event, _ := repo.Get(new_event.ID)
	expectedStatus := model.InProgress

	if event.Status != expectedStatus {
		t.Errorf("Event status. Expected: %v, Found: %v", expectedStatus, event.Status)
	}
}

func TestLockZero(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	lockedEvents, err := repo.Lock(0)
	expected := 0
	if len(lockedEvents) != expected {
		t.Errorf("Locked events count. Expected: %d, Found: %d", expected, len(lockedEvents))
	}
	expectedError := ErrLockPositive

	if err != expectedError {
		t.Errorf("Event lock error. Expected: %v, Found: %v", expectedError, err)
	}
}

func TestLockGreaterThanExists(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	lockedEvents, _ := repo.Lock(2)
	expected := 1

	if len(lockedEvents) != expected {
		t.Errorf("Locked events count. Expected: %d, Found: %d", expected, len(lockedEvents))
	}
}

func TestLockExactCount(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)
	new_event2 := createEvent(2, model.Created, model.Deferred, prod)
	repo.Add(*new_event2)

	lockedEvents, _ := repo.Lock(1)
	expected := 1

	if len(lockedEvents) != expected {
		t.Errorf("Locked events count. Expected: %d, Found: %d", expected, len(lockedEvents))
	}
}

func TestUnlock(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	lockedEvents, _ := repo.Lock(1)
	err := repo.Unlock([]uint64{lockedEvents[0].ID})

	if err != nil {
		t.Error(err.Error())
	}

	unlockedEvent, _ := repo.Get(new_event.ID)
	expectedStatus := model.Deferred

	if unlockedEvent.Status != expectedStatus {
		t.Errorf("Event status. Expected: %d, Found: %d", expectedStatus, unlockedEvent.Status)

	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	repo := NewInMemoryProductEventRepo()
	prod := createProduct(1)
	new_event := createEvent(1, model.Created, model.Deferred, prod)
	repo.Add(*new_event)

	err := repo.Remove([]uint64{1})
	if err != nil {
		t.Error(err.Error())
	}

	expectedCount := 0

	if int(repo.Count()) != expectedCount {
		t.Errorf("Events count. Expected: %d, Found: %d", expectedCount, repo.Count())

	}
}

func TestLock_ReacquireStuckInProgress(t *testing.T) {
	repo := &InMemoryProductEventRepo{
		events:            map[uint64]*model.ProductEvent{},
		processingTimeout: 1, // 1 секунда для теста
	}

	// Добавляем событие со статусом InProgress, которое "зависло"
	event := &model.ProductEvent{
		ID:           42,
		Type:         model.Created,
		Status:       model.InProgress,
		ProcessingAt: time.Now().Unix() - 2, // прошло больше таймаута
	}
	repo.Add(*event)

	locked, err := repo.Lock(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locked) != 1 {
		t.Fatalf("expected 1 event, got %d", len(locked))
	}
	if locked[0].ID != event.ID {
		t.Errorf("expected event ID %d, got %d", event.ID, locked[0].ID)
	}
	if locked[0].ProcessingAt <= event.ProcessingAt {
		t.Errorf("expected updated ProcessingAt, got %d", locked[0].ProcessingAt)
	}
	if locked[0].Status != model.InProgress {
		t.Errorf("expected status InProgress, got %v", locked[0].Status)
	}
}
