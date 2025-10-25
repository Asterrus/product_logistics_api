package repo

import (
	"log"
	eventstorage "product_logistics_api/internal/app/event_storage"
	"product_logistics_api/internal/model"
)

// Часть Ретранслятора для работы с хранилищем событий
type EventRepo interface {
	// Получение n уникальных событий из хранилища и проставление
	//  у них в хранилище статуса "Обрабатывается"
	Lock(n uint64) ([]model.ProductEvent, error)
	// Смена статуса с "Обрабатывается" на "К обработке" по списку eventIDs
	Unlock(eventIDs []uint64) error
	// Создать модель события
	Add(event model.ProductEvent) error
	// Удалить обработанное событие?
	Remove(eventIDs []uint64) error
}

type ProductEventRepo struct {
	storage eventstorage.EventStorage
}

func (r *ProductEventRepo) Lock(n uint64) ([]model.ProductEvent, error) {
	log.Printf("ProductEventRepo Lock: %v", n)
	return r.storage.Lock(n)
}
func (r *ProductEventRepo) Unlock(eventIDs []uint64) error {
	log.Printf("ProductEventRepo Unlock eventIDs: %v", eventIDs)
	return nil
}
func (r *ProductEventRepo) Add(event model.ProductEvent) error {
	log.Printf("ProductEventRepo Add event: %v", event)
	return r.storage.Add(event)
}
func (r *ProductEventRepo) Remove(eventIDs []uint64) error {
	log.Printf("ProductEventRepo Remove eventIDs: %v", eventIDs)
	return nil
}

func NewProductEventRepo() EventRepo {
	return &ProductEventRepo{
		storage: eventstorage.NewInMemoryEventStorage(),
	}
}
