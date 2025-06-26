package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Resultado struct {
	Tarifa  string // se leerá como string desde el hash
	Latency string
	Worker  string
}

func main() {
	// 1. Conexión a Redis ---------------------------------------------
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	const cantidad = 10 // número de jobs enviados con push_job.go
	fmt.Printf("⌛ Esperando %d resultados...\n", cantidad)

	var resultados []Resultado
	inicio := time.Now()

	for i := 1; i <= cantidad; i++ {
		clave := fmt.Sprintf("tarifa:result:job%d", i)

		for {
			// Verificamos si la clave existe
			ex, err := rdb.Exists(ctx, clave).Result()
			if err != nil {
				log.Fatal(err)
			}
			if ex == 0 {
				time.Sleep(80 * time.Millisecond) // aún no procesado
				continue
			}

			// Recuperamos el hash completo
			val, err := rdb.HGetAll(ctx, clave).Result()
			if err != nil {
				log.Fatal(err)
			}

			// Aseguramos que tenga los campos esperados
			if len(val) == 0 {
				time.Sleep(40 * time.Millisecond)
				continue
			}

			res := Resultado{
				Tarifa:  val["tarifa"],
				Latency: val["latency_ns"],
				Worker:  val["worker_id"],
			}
			resultados = append(resultados, res)
			break
		}
	}

	duracion := time.Since(inicio)
	fmt.Printf("🎯 Todos los resultados listos en %.2f s\n", duracion.Seconds())

	// 2. Mostrar en consola -------------------------------------------
	for i, res := range resultados {
		fmt.Printf("➜ job%02d  tarifa=%s  latency=%s ns  worker=%s\n",
			i+1, res.Tarifa, res.Latency, res.Worker)
	}

	// 3. Guardar en CSV -----------------------------------------------
	// crea carpeta results si no existe
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
