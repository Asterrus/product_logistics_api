package retranslator

import (
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/app/sender"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	repo := repo.NewProductEventRepo()
	sender := sender.NewProductEventSender()

	cfg := Config{
		ChannelSize:    512,
		ConsumerCount:  2,
		ConsumeSize:    10,
		ConsumeTimeout: 10 * time.Second,
		ProducerCount:  2,
		WorkerCount:    2,
		Repo:           repo,
		Sender:         sender,
	}

	retranslator := NewRetranslator(cfg)
	retranslator.Start()
	retranslator.Close()
}

func TestConsumer(t *testing.T) {
	repo := repo.NewProductEventRepo()
	sender := sender.NewProductEventSender()

	cfg := Config{
		ChannelSize:    512,
		ConsumerCount:  2,
		ConsumeSize:    10,
		ConsumeTimeout: 10 * time.Second,
		ProducerCount:  2,
		WorkerCount:    2,
		Repo:           repo,
		Sender:         sender,
	}

	retranslator := NewRetranslator(cfg)
	retranslator.Start()
	retranslator.Close()
}
