package sender

import (
	"log"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
)

type InMemoryProductEventSender struct {
	sendsCount uint64
}

type TestEventSender interface {
	ports.EventSender
	SendsCount() uint64
}

func NewProductEventSender() TestEventSender {
	return &InMemoryProductEventSender{
		sendsCount: uint64(0),
	}
}

func (s *InMemoryProductEventSender) Send(product *model.ProductEvent) error {
	log.Printf("ProductEventSender Unlock product: %v", product)
	s.sendsCount++
	return nil
}

func (s *InMemoryProductEventSender) SendsCount() uint64 {
	log.Printf("ProductEventSender SendsCount")
	return s.sendsCount
}
