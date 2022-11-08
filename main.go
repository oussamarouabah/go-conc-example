package main

import (
	"fmt"
	"sync"
)

// Generator
func Generator() <-chan int {
	ch := make(chan int, 5)
	go func() {
		for i := 0; i < 1000; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// worker
func worker(ch <-chan int, out1, out2 chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range ch {
		if i%2 == 0 {
			out1 <- i
		}
		out2 <- i
	}
}

func results(out1, out2 <-chan int, done chan struct{}) {
	for {
		select {
		case i, ok := <-out1:
			if !ok {
				out1 = nil
				continue
			}
			fmt.Printf("%v is even\n", i)
		case i, ok := <-out2:
			if !ok {
				out2 = nil
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
		wg sync.WaitGroup

		out1 = make(chan int, 5)
		out2 = make(chan int, 5)
		done = make(chan struct{})
	)

	ch := Generator()

	wg.Add(4)
	go worker(ch, out1, out2, &wg)
	go worker(ch, out1, out2, &wg)
	go worker(ch, out1, out2, &wg)
	go worker(ch, out1, out2, &wg)

	go func() {
		wg.Wait()
		close(out1)
		close(out2)
		done <- struct{}{}
	}()

	results(out1, out2, done)
}
