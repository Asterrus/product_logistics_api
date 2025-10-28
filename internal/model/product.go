package model

import "github.com/google/uuid"

type Product struct {
	ID uint64
}

type EventType uint8

type EventStatus uint8

type EventProcessedResult uint8

const (
	Created EventType = iota
	Updated
	Removed
)

const (
	Deferred EventStatus = iota
	InProgress
	Processed // Сейчас не используется, так как удаляется событие при успешной отправке, но возможно не стоит удалять?
)

const (
	Sent EventProcessedResult = iota
	Returned
)

type ProductEvent struct {
	ID     uint64
	Type   EventType
	Status EventStatus
	Entity *Product
}

type ProductEventProcessed struct {
	ID      uuid.UUID
	EventID uint64
	Result  EventProcessedResult
}
