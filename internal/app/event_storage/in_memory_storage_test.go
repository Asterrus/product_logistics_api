package eventstorage

import (
	"product_logistics_api/internal/model"
	"testing"
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
	storage := NewInMemoryEventStorage()
	product := createProduct(1)
	event := createEvent(1, model.Created, model.Deffered, product)
	storage.Add(*event)
	if count := storage.Сount(); count != 1 {
		t.Errorf("expected 1 event, found: %d", count)
	}
}

func TestLock(t *testing.T) {
	storage := NewInMemoryEventStorage()
	product := createProduct(1)
	event := createEvent(1, model.Created, model.Deffered, product)
	storage.Add(*event)
	events, _ := storage.Lock(1)
	if len(events) != 1 {
		t.Errorf("expected 1 event, found: %d", len(events))
	}
}
