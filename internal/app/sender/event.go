package sender

import (
	"product_logistics_api/internal/model"
)

// Часть Ретранслятора для работы с брокером сообщений
type EventSender interface {
	// Отправить сообщение в брокер
	Send(product *model.ProductEvent) error
}

type TestEventSender interface {
	EventSender
	SendsCount() uint64
}
