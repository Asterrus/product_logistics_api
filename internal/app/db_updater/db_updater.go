package dbupdater

import (
	"context"
	"log"
	"product_logistics_api/internal/model"
	"product_logistics_api/internal/ports"
	"sync"
	"time"
)

type dbUpdater struct {
	updatersCount   uint64
	processedEvents <-chan model.ProductEventProcessed
	wg              *sync.WaitGroup
	timeout         time.Duration
	batchSize       uint64
	repo            ports.EventRepo
}

func NewDbUpdater(
	updatersCount uint64,
	processedEvents <-chan model.ProductEventProcessed,
	timeout time.Duration,
	batchSize uint64,
	repo ports.EventRepo,
) ports.DbUpdater {
	wg := &sync.WaitGroup{}

	return &dbUpdater{
		updatersCount:   updatersCount,
		processedEvents: processedEvents,
		wg:              wg,
		timeout:         timeout,
		batchSize:       batchSize,
		repo:            repo,
	}
}

func (u *dbUpdater) Start(ctx context.Context) {
	for i := 0; i < int(u.updatersCount); i++ {
		u.wg.Add(1)
		go func() {
			defer u.wg.Done()
			ticker := time.NewTicker(u.timeout)
			defer ticker.Stop()

			for {
				events := []model.ProductEventProcessed{}
				loop := true
				for i := 0; (i < int(u.batchSize)) && loop; i++ {
					select {
					case <-ctx.Done():
						log.Println("DbUpdater stopped by context:", ctx.Err())
						return
					case e := <-u.processedEvents:
						log.Printf("Find element %v of %v total", i+1, u.batchSize)
						events = append(events, e)
					case <-ticker.C:
						log.Println("DbUpdater case ticker.C")
						loop = false
					}
				}
				if len(events) == 0 {
					continue
				}
				log.Printf("dbUpdater. %v events to handle", len(events))
				for _, e := range events {
					if e.Result == model.Sent {

						if err := u.repo.Remove([]uint64{e.EventID}); err != nil {
							// Что делаем если не удалось удалить запись о событии из базы?
							log.Printf("Repo Remove error: %v", err)
						}

					} else {

						if err := u.repo.Unlock([]uint64{e.EventID}); err != nil {
							// Что делаем если не вышло вернуть записи статус "К обработке"?
							log.Printf("Repo Unlock error: %v", err)
						}
					}
				}

			}

		}()
	}
}
func (u *dbUpdater) Close() {
	u.wg.Wait()
}
