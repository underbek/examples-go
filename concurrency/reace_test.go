package concurrency

import (
	"fmt"
	"sync"
	"testing"
)

func TestRace(_ *testing.T) {

	results := make(map[string]int)

	for i := 0; i < 10_000_000; i++ {
		var a, b int

		wg := sync.WaitGroup{}
		wg.Add(2)

		done := make(chan struct{}, 1)

		go func() {

			a = 1
			b = 2
			done <- struct{}{}

			wg.Done()
		}()

		go func() {

			<-done
			c := a
			d := b

			key := fmt.Sprintf("%d:%d", c, d)
			results[key]++

			wg.Done()
		}()

		wg.Wait()
	}

	for k, v := range results {
		fmt.Printf("[%s] - %d\n", k, v)
	}
}
