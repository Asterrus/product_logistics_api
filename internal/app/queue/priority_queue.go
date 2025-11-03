package queue

import (
	"container/heap"
	"product_logistics_api/internal/model"
	"sync"
)

type EventQueue interface {
	heap.Interface
	PushEvent(*model.ProductEvent)
	PopEvent() *model.ProductEvent
}
type productEvents struct {
	mu    *sync.Mutex
	items []*model.ProductEvent
}

// type productEvents []*model.ProductEvent

func (q productEvents) Len() int {
	return len(q.items)
}

func (q productEvents) Less(i, j int) bool {
	if q.items[i].Type == q.items[j].Type {
		return q.items[i].ID < q.items[j].ID
	}
	return q.items[i].Type < q.items[j].Type
}

func (q productEvents) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *productEvents) Push(event any) {
	e := event.(*model.ProductEvent)
	q.items = append(q.items, e)

}

func (q *productEvents) Pop() any {
	old := q.items
	n := len(old)
	item := old[n-1]
	q.items = old[0 : n-1]
	return item
}

func (q *productEvents) PopEvent() *model.ProductEvent {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	return heap.Pop(q).(*model.ProductEvent)
}

func (q *productEvents) PushEvent(event *model.ProductEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(q, event)
}
func (q productEvents) Peek() *model.ProductEvent {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}
func NewProductQueue(events []*model.ProductEvent) EventQueue {
	if events == nil {
		events = []*model.ProductEvent{}
	}
	mu := &sync.Mutex{}
	q := &productEvents{
		items: events,
		mu:    mu,
	}
	heap.Init(q)
	return q
}

var _ EventQueue = (*productEvents)(nil)
