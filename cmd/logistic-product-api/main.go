package main

import (
	"fmt"
	"sync"

	"github.com/gammazero/workerpool"
)

func test() {
	wp := workerpool.New(2)
	requests := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	for _, r := range requests {
		r := r
		wp.Submit(func() {
			fmt.Println("Handling request:", r)
		})
	}

	wp.StopWait()
}
func test2() {
	type WorkerPool struct {
		closed bool
		mu     sync.Mutex

		tasks     chan func()
		wgWorkers sync.WaitGroup
		doneCh    chan struct{}
	}
	wp := WorkerPool{}
	fmt.Println(wp)
	fmt.Printf("%v\n", wp.closed)
	fmt.Printf("%v\n", wp.mu)
	fmt.Printf("%v\n", wp.tasks)
	fmt.Printf("%v\n", wp.doneCh)
	fmt.Printf("%v\n", wp.wgWorkers)
}
func main() {

	// sigs := make(chan os.Signal, 1)

	// cfg := retranslator.Config{
	// 	ChannelSize:    512,
	// 	ConsumerCount:  2,
	// 	ConsumeSize:    10,
	// 	ProducerCount:  28,
	// 	WorkerCount:    2,
	// 	ConsumeTimeout: 2,
	// }

	// retranslator := retranslator.NewRetranslator(cfg)
	// retranslator.Start()

	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// <-sigs
	test2()
}
