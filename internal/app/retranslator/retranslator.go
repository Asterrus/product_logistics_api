package retranslator

import (
	"context"
	"product_logistics_api/internal/app/consumer"
	"product_logistics_api/internal/app/producer"
	"product_logistics_api/internal/app/workerpool"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"time"
)

type Retranslator interface {
	Start(context.Context)
	Close()
}

type Config struct {
	EventsChannelSize          uint64
	ProcessedEventsChannelSize uint64

	ConsumerCount  uint64
	ConsumeSize    uint64
	ConsumeTimeout time.Duration

	ProducerCount        uint64
	WorkerCount          int
	WorkerPoolBufferSize uint64

	Repo   ports.EventRepo
	Sender ports.EventSender
}

type retranslator struct {
	events     chan model.ProductEvent
	consumer   consumer.Consumer
	producer   producer.Producer
	workerPool ports.TaskSubmitter
}

func NewRetranslator(cfg Config) Retranslator {
	events := make(chan model.ProductEvent, cfg.EventsChannelSize)
	workerPool, _ := workerpool.NewWorkerPool(uint64(cfg.WorkerCount), uint64(cfg.WorkerPoolBufferSize))
	processedEventsChannel := make(chan model.ProductEventProcessed, cfg.ProcessedEventsChannelSize)
	consumer := consumer.NewDbConsumer(
		cfg.ConsumerCount,
		cfg.ConsumeSize,
		cfg.ConsumeTimeout,
		cfg.Repo,
		events,
		processedEventsChannel)
	producer := producer.NewKafkaProducer(
		cfg.ProducerCount,
		cfg.Sender,
		events,
		processedEventsChannel,
		workerPool)

	return &retranslator{
		events:     events,
		consumer:   consumer,
		producer:   producer,
		workerPool: workerPool,
	}
}

func (r *retranslator) Start(ctx context.Context) {
	r.producer.Start(ctx)
	r.consumer.Start(ctx)
}

func (r *retranslator) Close() {
	r.consumer.Close()
	r.producer.Close()
	r.workerPool.StopWait()
}
