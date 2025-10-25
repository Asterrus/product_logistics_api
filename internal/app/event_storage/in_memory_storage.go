package eventstorage

import (
	"log"
	"product_logistics_api/internal/model"
	"sync"
)

type InMemoryEventStorage struct {
	events map[uint64]*model.ProductEvent
	mu     sync.Mutex
}

func NewInMemoryEventStorage() EventStorage {
	return &InMemoryEventStorage{
		events: map[uint64]*model.ProductEvent{},
	}
}

func (r *InMemoryEventStorage) Lock(n uint64) ([]model.ProductEvent, error) {
	log.Printf("InMemoryEventStorage Lock n: %v", n)
	r.mu.Lock()
	defer r.mu.Unlock()
	res := []model.ProductEvent{}

	for _, e := range r.events {
		if e.Status == model.Deffered {
			res = append(res, *e)
			e.Status = model.InProgress
		}
	}
	return res, nil
}

func (r *InMemoryEventStorage) Add(event model.ProductEvent) error {
	log.Printf("InMemoryEventStorage Add event: %v", event)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.ID] = &event
	return nil
}

func (r *InMemoryEventStorage) Remove(eventIDs []uint64) error {
	log.Printf("InMemoryEventStorage Remove: %v", eventIDs)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range eventIDs {
		delete(r.events, id)
	}
	return nil
}
func (r *InMemoryEventStorage) Unlock(eventIDs []uint64) error {
	log.Printf("InMemoryEventStorage Unlock: %v", eventIDs)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range eventIDs {
		value, ok := r.events[id]
		if ok {
			value.Status = model.Deffered
		}
	}
	return nil
}

func (r *InMemoryEventStorage) Сount() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return uint64(len(r.events))
}
