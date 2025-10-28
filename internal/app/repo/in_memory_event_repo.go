package repo

import (
	"errors"
	"fmt"
	"log"
	"product_logistics_api/internal/model"
	"sync"
)

var (
	ErrLockPositive = errors.New("Lock argument must be positive")
)

type InMemoryProductEventRepo struct {
	events    map[uint64]*model.ProductEvent
	mu        sync.Mutex
	lockCalls uint64
}

func NewInMemoryProductEventRepo() TestEventRepo {
	return &InMemoryProductEventRepo{
		events: map[uint64]*model.ProductEvent{},
	}
}

func (r *InMemoryProductEventRepo) Lock(n uint64) ([]model.ProductEvent, error) {
	log.Printf("InMemoryProductEventRepo Lock: %v", n)

	if n <= 0 {
		return nil, ErrLockPositive
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	res := []model.ProductEvent{}

	for _, e := range r.events {
		if e.Status == model.Deffered {
			e.Status = model.InProgress
			res = append(res, *e)

		}
	}
	r.lockCalls++
	return res, nil
}
func (r *InMemoryProductEventRepo) Unlock(eventIDs []uint64) error {
	log.Printf("InMemoryProductEventRepo Unlock eventIDs: %v", eventIDs)
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
func (r *InMemoryProductEventRepo) Add(event model.ProductEvent) error {
	log.Printf("InMemoryProductEventRepo Add event: %v", event)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.ID] = &event
	return nil
}
func (r *InMemoryProductEventRepo) Remove(eventIDs []uint64) error {
	log.Printf("InMemoryProductEventRepo Remove eventIDs: %v", eventIDs)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range eventIDs {
		delete(r.events, id)
	}
	return nil
}
func (r *InMemoryProductEventRepo) Count() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return uint64(len(r.events))
}

func (r *InMemoryProductEventRepo) Get(eventID uint64) (model.ProductEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	event, ok := r.events[eventID]
	if !ok {
		return model.ProductEvent{}, fmt.Errorf("Event with ID %d not found", eventID)
	}
	return *event, nil
}

func (r *InMemoryProductEventRepo) CountOfLocksCall() uint64 {
	return r.lockCalls
}
