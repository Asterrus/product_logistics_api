package ports

import "product_logistics_api/internal/model"

// Хранилище событий
type EventRepo interface {
	Lock(n uint64) ([]model.ProductEvent, error)
	Unlock(eventIDs []uint64) error
	Add(event model.ProductEvent) error
	Remove(eventIDs []uint64) error
}

// Отправка событий во внешний брокер
type EventSender interface {
	Send(product *model.ProductEvent) error
}

// Минимальный контракт пула задач
type TaskSubmitter interface {
	Submit(func())
}
