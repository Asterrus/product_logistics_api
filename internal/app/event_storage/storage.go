package eventstorage

import "product_logistics_api/internal/model"

type EventStorage interface {
	Lock(n uint64) ([]model.ProductEvent, error)
	Add(event model.ProductEvent) error
	Unlock(eventIDs []uint64) error
	Remove(eventIDs []uint64) error
	Сount() uint64
}
