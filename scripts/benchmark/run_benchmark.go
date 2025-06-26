package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID      string  `json:"id"`
	Consumo float64 `json:"consumo"`
	Uso     int     `json:"uso"`
	Grupo   int     `json:"grupo"`
	Empresa int     `json:"empresa"`
}

type Resultado struct {
	Tarifa  int
	Latency int64
	Worker  string
}

func main() {
	const cantidad = 10
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	fmt.Println("🚀 Iniciando benchmark...")

	// 1. Enviar jobs
	for i := 1; i <= cantidad; i++ {
		job := Job{
			ID:      fmt.Sprintf("job%d", i),
			Consumo: float64(100 + i*10),
			Uso:     i % 3,
			Grupo:   i % 2,
			Empresa: i % 5,
		}
		payload, _ := json.Marshal(job)
		rdb.LPush(ctx, "tarifa:jobs", payload)
	}

	// 2. Esperar resultados
	fmt.Printf("⌛ Esperando %d resultados...\n", cantidad)
	var resultados []Resultado
	inicio := time.Now()

	for i := 1; i <= cantidad; i++ {
		clave := fmt.Sprintf("tarifa:result:job%d", i)

		for {
			val, err := rdb.HGetAll(ctx, clave).Result()
			if err != nil {
				log.Fatal(err)
			}
			if len(val) == 0 {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			latency, _ := strconv.ParseInt(val["latency_ns"], 10, 64)
			tarifa, _ := strconv.Atoi(val["tarifa"])
			worker := val["worker_id"]

			res := Resultado{
				Tarifa:  tarifa,
				Latency: latency,
				Worker:  worker,
			}
			resultados = append(resultados, res)
			break
		}
	}

	tiempoTotal := time.Since(inicio)
	fmt.Printf("✅ Benchmark terminado en %.2f segundos\n", tiempoTotal.Seconds())

	// 3. Guardar resultados
	file, err := os.Create("results/benchmark.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"job_id", "tarifa", "latency_ns", "worker"})
	for i, res := range resultados {
		writer.Write([]string{
			fmt.Sprintf("job%d", i+1),
			fmt.Sprintf("%d", res.Tarifa),
			fmt.Sprintf("%d", res.Latency),
			res.Worker,
		})
	}
	fmt.Println("📊 Resultados guardados en results/benchmark.csv")
}
