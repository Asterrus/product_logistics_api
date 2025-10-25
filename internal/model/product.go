package model

type Product struct {
	ID uint64
}

type EventType uint8

type EventStatus uint8

const (
	Created EventType = iota
	Updated
	Removed
)

const (
	Deffered EventStatus = iota
	InProgress
	Processed
)

type ProductEvent struct {
	ID     uint64
	Type   EventType
	Status EventStatus
	Entity *Product
}
