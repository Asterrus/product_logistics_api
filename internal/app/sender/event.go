package sender

import (
	"log"
	"product_logistics_api/internal/model"
)

// Часть Ретранслятора для работы с брокером сообщений
type EventSender interface {
	// Отправить сообщение в брокер
	Send(product *model.ProductEvent) error
}

type ProductEventSender struct{}

func (s *ProductEventSender) Send(product *model.ProductEvent) error {
	log.Printf("ProductEventSender Unlock product: %v", product)
	return nil
}

func NewProductEventSender() EventSender {
	return &ProductEventSender{}
}
