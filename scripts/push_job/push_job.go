package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
	// ► Lee REDIS_ADDR (ej. "redis:6379"); fallback a localhost fuera de Docker
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx := context.Background()

	for i := 1; i <= 10; i++ {
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
