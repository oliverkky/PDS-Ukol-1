package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

const numThreads = 1_024

var globalCounter atomic.Int64

var pole [numThreads]atomic.Int64
var vstup [numThreads]atomic.Bool

func main() {
	for i := range numThreads { // Pustil jsem `go fix lamportsbakery.go` a přidalo to syntaktický cukr do této smyčky
		go lamportsBakery(i)
	}

	// Udržet main naživu, dokud nedoběhnou všechna vlákna
	for globalCounter.Load() < int64(numThreads) {
		runtime.Gosched()
	}

	expected := int64(numThreads)
	fmt.Printf("All workers finished. Final counter = %d (expected %d)\n",
		globalCounter.Load(), expected)
}

func lamportsBakery(i int) {
	// 1. nekritická sekce
	fmt.Printf("Non-critical section of thread %3d.\n", i)

	// 2. vstup[i] <- true
	vstup[i].Store(true)

	// 3. pole[i] <- max(pole) + 1
	var maxTicket int64
	for j := range numThreads {
		if value := pole[j].Load(); value > maxTicket {
			maxTicket = value
		}
	}
	pole[i].Store(maxTicket + 1)

	// 4. vstup[i] <- false
	vstup[i].Store(false)

	// 5. pro všechna j různá od i
	for j := range numThreads {
		if j == i {
			continue
		}

		// await(not(vstup[j]))
		for vstup[j].Load() {
			runtime.Gosched()
		}

		// await(pole[j] = 0 or pole[i] < pole[j] oatomic.Int6r
		//       (pole[i] = pole[j] and i < j))
		for {
			jTicket := pole[j].Load()
			iTicket := pole[i].Load()
			if jTicket == 0 || iTicket < jTicket ||
				(iTicket == jTicket && i < j) {
				break
			}
			runtime.Gosched()
		}
	}

	// 6. kritická sekce
	globalCounter.Store(globalCounter.Load() + 1)
	fmt.Printf("Thread %3d incremented global counter to %5d.\n", i, globalCounter.Load())

	// 7. pole[i] <- 0
	pole[i].Store(0)
}
