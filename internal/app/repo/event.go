package repo

import (
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
