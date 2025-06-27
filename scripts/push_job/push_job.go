package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID      string  `json:"id"`
	Consumo float64 `json:"consumo"`
	Uso     int     `json:"uso"`
	Grupo   int     `json:"grupo"`
	Empresa int     `json:"empresa"`
}

func main() {
	// ► REDIS_ADDR (p.e. "redis:6379"); por defecto localhost
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	// ► CANTIDAD de jobs ─ lee variable de entorno; default 10
	numJobs := 10
	if s := os.Getenv("CANTIDAD"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			numJobs = n
		}
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx := context.Background()

	for i := 1; i <= numJobs; i++ {
		job := Job{
			ID:      fmt.Sprintf("job%d", i),
			Consumo: float64(100 + i*10),
			Uso:     i % 3,
			Grupo:   i % 2,
			Empresa: i % 5,
		}

		payload, err := json.Marshal(job)
		if err != nil {
			log.Fatal(err)
		}

		if err := rdb.LPush(ctx, "tarifa:jobs", payload).Err(); err != nil {
			log.Fatal(err)
		}
		log.Printf("🚀 Enviado %s", job.ID)
	}
}
