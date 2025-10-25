package consumer

import (
	"fmt"
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/model"
	"testing"
	"time"
)

func TestNewDbConsumer(t *testing.T) {
	repo := repo.NewProductEventRepo()
	eventsChannelBuffer := 2
	eventsChannel := make(chan model.ProductEvent, eventsChannelBuffer)
	batchSize := uint64(1)
	consumersNumber := uint64(1)
	cunsumeTimeout := time.Second
	consumer := NewDbConsumer(consumersNumber, batchSize, cunsumeTimeout, repo, eventsChannel)
	fmt.Printf("Consumer : %v\n", consumer)
	consumer.Start()
	time.Sleep(time.Second * 5)
}
