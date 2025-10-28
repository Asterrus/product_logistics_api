package main

import (
	"os"
	"os/signal"
	"product_logistics_api/internal/app/repo"
	"product_logistics_api/internal/app/retranslator"
	"product_logistics_api/internal/app/sender"
	"syscall"
	"time"
)

func main() {
	sigs := make(chan os.Signal, 1)
	repo := repo.NewInMemoryProductEventRepo()
	sender := sender.NewProductEventSender()
	cfg := retranslator.Config{
		EventsChannelSize:          512,
		ProcessedEventsChannelSize: 512,
		ConsumerCount:              2,
		ConsumeSize:                10,
		ProducerCount:              28,
		WorkerCount:                2,
		ConsumeTimeout:             1 * time.Second,
		Repo:                       repo,
		Sender:                     sender,
	}

	retranslator := retranslator.NewRetranslator(cfg)
	retranslator.Start()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}
