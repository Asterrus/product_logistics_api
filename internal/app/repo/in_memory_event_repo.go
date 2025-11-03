package repo

import (
	"errors"
	"fmt"
	"log"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"sync"
	"time"
)

var (
	ErrLockPositive = errors.New("Lock argument must be positive")
)

type InMemoryProductEventRepo struct {
	events            map[uint64]*model.ProductEvent
	mu                sync.Mutex
	lockCalls         uint64
	processingTimeout int64 // сек
}

type TestEventRepo interface {
	ports.EventRepo
	// Для тестов
	Count() uint64

	Get(eventID uint64) (model.ProductEvent, error)

	CountOfLocksCall() uint64
}

func NewInMemoryProductEventRepo() TestEventRepo {
	return &InMemoryProductEventRepo{
		events:            map[uint64]*model.ProductEvent{},
		processingTimeout: 10, // default, можно передавать через конструктор
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
	found := 0
	now := getUnixNow()
	for _, e := range r.events {
		if e.Status == model.Deferred {
			e.Status = model.InProgress
			e.ProcessingAt = now
			res = append(res, *e)
			found++
		} else if e.Status == model.InProgress && (now-e.ProcessingAt) > r.processingTimeout {
			// зависшее событие, повторно берём в обработку
			e.ProcessingAt = now
			res = append(res, *e)
			found++
		}
		if found == int(n) {
			break
		}
	}
	r.lockCalls++
	return res, nil

}

func getUnixNow() int64 {
	return time.Now().Unix()
}
func (r *InMemoryProductEventRepo) Unlock(eventIDs []uint64) error {
	log.Printf("InMemoryProductEventRepo Unlock eventIDs: %v", eventIDs)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range eventIDs {
		value, ok := r.events[id]
		if ok {
			value.Status = model.Deferred
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
		return model.ProductEvent{}, fmt.Errorf("event with ID %d not found", eventID)
	}
	return *event, nil
}

func (r *InMemoryProductEventRepo) CountOfLocksCall() uint64 {
	return r.lockCalls
}
