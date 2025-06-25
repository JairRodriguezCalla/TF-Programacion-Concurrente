package main

import (
	"context"
	"encoding/json"
	"log"

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
	// Conexión al mismo Redis del contenedor (puerto 6379 en host)
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // ⚠️ usa IPv4 explícita
	})

	job := Job{
		ID:      "demo1",
		Consumo: 300,
		Uso:     1,
		Grupo:   2,
		Empresa: 7,
	}

	payload, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("error al serializar job: %v", err)
	}

	ctx := context.Background()
	if err := rdb.LPush(ctx, "tarifa:jobs", payload).Err(); err != nil {
		log.Fatalf("error al enviar job a Redis: %v", err)
	}

	log.Println("✅ Job enviado correctamente a Redis")
}
