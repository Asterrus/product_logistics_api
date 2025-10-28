package sender

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
func TestSend(t *testing.T) {
	sender := NewProductEventSender()
	sender.Send(createEvent(1, model.Created, model.Deffered, createProduct(1)))
	expected := 1
	if sender.SendsCount() != 1 {
		t.Errorf("Sends count must be %v", expected)
	}
}
