package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Resultado struct {
	Tarifa  string
	Latency string
	Worker  string
}

func main() {
	// ► Lee REDIS_ADDR; fallback a localhost fuera de Docker
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx := context.Background()

	cantidad := 10
	if val := os.Getenv("CANTIDAD"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			cantidad = n
		}
	}
	fmt.Printf("⌛ Esperando %d resultados...\n", cantidad)

	var resultados []Resultado
	inicio := time.Now()

	for i := 1; i <= cantidad; i++ {
		clave := fmt.Sprintf("tarifa:result:job%d", i)

		for {
			ex, err := rdb.Exists(ctx, clave).Result()
			if err != nil {
				log.Fatal(err)
			}
			if ex == 0 {
				time.Sleep(80 * time.Millisecond)
				continue
			}

			val, err := rdb.HGetAll(ctx, clave).Result()
			if err != nil {
				log.Fatal(err)
			}
			if len(val) == 0 {
				time.Sleep(40 * time.Millisecond)
				continue
			}

			resultados = append(resultados, Resultado{
				Tarifa:  val["tarifa"],
				Latency: val["latency_ns"],
				Worker:  val["worker_id"],
			})
			break
		}
	}

	fmt.Printf("🎯 Todos los resultados listos en %.2f s\n", time.Since(inicio).Seconds())

	// Mostrar en consola
	for i, res := range resultados {
		fmt.Printf("➜ job%02d  tarifa=%s  latency=%s ns  worker=%s\n",
			i+1, res.Tarifa, res.Latency, res.Worker)
	}

	// Guardar en CSV
	_ = os.MkdirAll("results", 0o755)
	f, err := os.Create("results/resultados.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"job_id", "tarifa", "latency_ns", "worker"})
	for i, res := range resultados {
		w.Write([]string{
			fmt.Sprintf("job%d", i+1),
			res.Tarifa,
			res.Latency,
			res.Worker,
		})
	}

	fmt.Println("✅ Resultados guardados en results/resultados.csv")
}
