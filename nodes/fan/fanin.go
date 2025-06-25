package fan

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func predictBasic(c float64) string {
	if c < 100 {
		return "BT5A"
	} else if c < 500 {
		return "BT5B"
	}
	return "BT6"
}

func fan(chunk []float64, out chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, v := range chunk {
		out <- predictBasic(v)
	}
}

func Demo() {
	rand.Seed(time.Now().UnixNano())
	const total = 40
	const fans = 4

	values := make([]float64, total)
	for i := range values {
		values[i] = rand.Float64()*700 + 10
	}

	chunk := (total + fans - 1) / fans
	out := make(chan string, total)

	var wg sync.WaitGroup
	for i := 0; i < fans; i++ {
		s := i * chunk
		e := s + chunk
		if s >= total {
			break
		}
		if e > total {
			e = total
		}
		wg.Add(1)
		go fan(values[s:e], out, &wg)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	count := map[string]int{}
	for r := range out {
		count[r]++
	}
	fmt.Println("Histograma Fan-in:")
	for k, v := range count {
		fmt.Printf("  %s : %d\n", k, v)
	}
}
