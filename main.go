package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Generator
func Generator(ctx context.Context) <-chan int {
	ch := make(chan int, 5)
	go func(ctx context.Context) {
		defer close(ch)
		for i := 0; i < 1000; i++ {
			select {
			case ch <- i:
			case <-ctx.Done():
				fmt.Println("Generator Ctx done with err =", ctx.Err().Error())
				return
			}
		}
	}(ctx)
	return ch
}

// worker
func worker(ctx context.Context, ch <-chan int, even, odd chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case i, ok := <-ch:
			time.Sleep(500 * time.Millisecond)
			if !ok {
				ch = nil
				continue
			}
			if i%2 == 0 {
				even <- i
			}
			odd <- i
		case <-ctx.Done():
			fmt.Println("Worker Ctx done with err = ", ctx.Err().Error())
			return
		}
	}
}

func results(even, odd <-chan int, done chan struct{}) {
	for {
		select {
		case i, ok := <-even:
			if !ok {
				even = nil
				continue
			}
			fmt.Printf("%v is even\n", i)
		case i, ok := <-odd:
			if !ok {
				odd = nil
				continue
			}
			fmt.Printf("%v is odd\n", i)
		case <-done:
			fmt.Println("Done printing results")
			return
		}
	}
}

func main() {
	var (
		wg        sync.WaitGroup
		ctx, stop = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

		even = make(chan int, 5)
		odd  = make(chan int, 5)
		done = make(chan struct{})
	)

	defer stop()
	ch := Generator(ctx)

	wg.Add(4)
	go worker(ctx, ch, even, odd, &wg)
	go worker(ctx, ch, even, odd, &wg)
	go worker(ctx, ch, even, odd, &wg)
	go worker(ctx, ch, even, odd, &wg)

	go func() {
		wg.Wait()
		close(even)
		close(odd)
		done <- struct{}{}
	}()

	results(even, odd, done)
}
