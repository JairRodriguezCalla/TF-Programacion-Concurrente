package workers

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type job struct {
	id      int
	consumo float64
}
type result struct {
	id     int
	tarifa string
}

func predictSimple(consumo float64) string {
	switch {
	case consumo < 100:
		return "BT5A"
	case consumo < 500:
		return "BT5B"
	default:
		return "BT6"
	}
}

func worker(id int, jobs <-chan job, out chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		out <- result{j.id, predictSimple(j.consumo)}
	}
}

func Demo() {
	rand.Seed(time.Now().UnixNano())
	const jobsN = 20
	const workers = 4

	jobs := make(chan job, jobsN)
	out := make(chan result, jobsN)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go worker(w, jobs, out, &wg)
	}

	for i := 1; i <= jobsN; i++ {
		jobs <- job{i, rand.Float64()*700 + 10}
	}
	close(jobs)

	wg.Wait()
	close(out)

	fmt.Println("Resultados Worker-Pool:")
	for r := range out {
		fmt.Printf("  job %02d → %s\n", r.id, r.tarifa)
	}
}
