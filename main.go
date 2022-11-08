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
func worker(ch <-chan int, even, odd chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range ch {
		if i%2 == 0 {
			even <- i
		}
		odd <- i
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
		wg sync.WaitGroup

		even = make(chan int, 5)
		odd  = make(chan int, 5)
		done = make(chan struct{})
	)

	ch := Generator()

	wg.Add(4)
	go worker(ch, even, odd, &wg)
	go worker(ch, even, odd, &wg)
	go worker(ch, even, odd, &wg)
	go worker(ch, even, odd, &wg)

	go func() {
		wg.Wait()
		close(even)
		close(odd)
		done <- struct{}{}
	}()

	results(even, odd, done)
}
