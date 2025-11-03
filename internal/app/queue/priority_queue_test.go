package queue

import (
	"product_logistics_api/internal/model"
	"testing"
)

func TestEmptyQueueReturnNil(t *testing.T) {
	q := NewProductQueue([]*model.ProductEvent{})
	if item := q.PopEvent(); item != nil {
		t.Error("Expected nil")
	}
}
func TestSingleProductEvents(t *testing.T) {
	p := &model.Product{ID: 1}
	events := []*model.ProductEvent{
		{ID: 1, Type: model.Created, Status: model.InProgress, Entity: p},
		{ID: 2, Type: model.Updated, Status: model.InProgress, Entity: p},
		{ID: 3, Type: model.Removed, Status: model.InProgress, Entity: p},
	}
	q := NewProductQueue([]*model.ProductEvent{events[2], events[0], events[1]})

	want := []model.EventType{model.Created, model.Updated, model.Removed}
	for i, expected := range want {
		got := q.PopEvent()
		if got.Type != expected {
			t.Errorf("[%d] expected type %v, got %v", i, expected, got.Type)
		}
	}
}
func TestTwoProductsEvents(t *testing.T) {

	p1 := &model.Product{ID: 1}
	p2 := &model.Product{ID: 2}
	events := []*model.ProductEvent{
		{ID: 1, Type: model.Created, Status: model.InProgress, Entity: p1},
		{ID: 2, Type: model.Updated, Status: model.InProgress, Entity: p1},
		{ID: 3, Type: model.Removed, Status: model.InProgress, Entity: p1},
		{ID: 4, Type: model.Created, Status: model.InProgress, Entity: p2},
		{ID: 5, Type: model.Updated, Status: model.InProgress, Entity: p2},
		{ID: 6, Type: model.Removed, Status: model.InProgress, Entity: p2},
	}

	q := NewProductQueue([]*model.ProductEvent{events[4], events[2], events[5], events[0], events[1], events[3]})

	tests := []struct {
		expectedType model.EventType
		expectedID   uint64
	}{
		{model.Created, 1},
		{model.Created, 4},
		{model.Updated, 2},
		{model.Updated, 5},
		{model.Removed, 3},
		{model.Removed, 6},
	}

	for i, tt := range tests {
		event := q.PopEvent()
		if expectedType := tt.expectedType; event.Type != expectedType {
			t.Errorf("[%d] expected event type: %v, found: %v", i, expectedType, event.Type)
		}
		if expectedId := tt.expectedID; event.ID != expectedId {
			t.Errorf("[%d] expected event id: %v, found: %v", i, expectedId, event.ID)
		}

	}
}
