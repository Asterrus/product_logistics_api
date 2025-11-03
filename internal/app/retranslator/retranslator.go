package retranslator

import (
	"context"
	"product_logistics_api/internal/app/consumer"
	dbupdater "product_logistics_api/internal/app/db_updater"
	"product_logistics_api/internal/app/producer"
	"product_logistics_api/internal/app/queue"
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
	ProcessedEventsChannelSize uint64

	ConsumerCount              uint64
	ConsumeSize                uint64
	ConsumeTimeout             time.Duration
	DbUpdatersTimeout          time.Duration
	DbUpdatersTimeoutBatchSize uint64

	ProducerCount        uint64
	ProducerTimeout      time.Duration
	WorkerCount          int
	DbUpdatersCount      uint64
	WorkerPoolBufferSize uint64

	Repo   ports.EventRepo
	Sender ports.EventSender
}

type retranslator struct {
	events     chan model.ProductEvent
	consumer   consumer.Consumer
	producer   producer.Producer
	dbUpdater  ports.DbUpdater
	workerPool ports.TaskSubmitter
}

func NewRetranslator(cfg Config) Retranslator {
	workerPool, _ := workerpool.NewWorkerPool(uint64(cfg.WorkerCount), uint64(cfg.WorkerPoolBufferSize))
	processedEventsChannel := make(chan model.ProductEventProcessed, cfg.ProcessedEventsChannelSize)
	queue := queue.NewProductQueue(nil)

	consumer := consumer.NewDbConsumer(
		cfg.ConsumerCount,
		cfg.ConsumeSize,
		cfg.ConsumeTimeout,
		cfg.Repo,
		processedEventsChannel,
		queue)
	producer := producer.NewKafkaProducer(
		cfg.ProducerCount,
		cfg.Sender,
		cfg.ProducerTimeout,
		processedEventsChannel,
		workerPool,
		queue)
	db_updater := dbupdater.NewDbUpdater(
		cfg.DbUpdatersCount,
		processedEventsChannel,
		cfg.DbUpdatersTimeout,
		cfg.DbUpdatersTimeoutBatchSize,
		cfg.Repo,
	)
	return &retranslator{
		consumer:   consumer,
		producer:   producer,
		workerPool: workerPool,
		dbUpdater:  db_updater,
	}
}

func (r *retranslator) Start(ctx context.Context) {
	r.producer.Start(ctx)
	r.consumer.Start(ctx)
	r.dbUpdater.Start(ctx)
}

func (r *retranslator) Close() {
	r.consumer.Close()
	r.producer.Close()
	r.dbUpdater.Close()

	r.workerPool.StopWait()
}
