package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"tf/nodes/rf"
	"tf/pkg/redisconn"
	"tf/pkg/tarifastats"
)

// ------------ estructura del Job ------------
type Job struct {
	ID      string  `json:"id"`
	Consumo float64 `json:"consumo"`
	Uso     int     `json:"uso"`
	Grupo   int     `json:"grupo"`
	Empresa int     `json:"empresa"`
}

// ------------ worker individual ------------
func worker(id int, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	workerName := os.Getenv("WORKER_ID")
	if workerName == "" {
		workerName = fmt.Sprintf("G%d", id)
	}

	rdb := redisconn.Client
	baseCtx := redisconn.Ctx // contexto base para Redis

	for {
		select {
		case <-ctx.Done():
			log.Printf("[W%d] 🛑 Cancelado por contexto", id)
			return
		default:
			// BRPOP bloquea hasta recibir algo
			res, err := rdb.BRPop(baseCtx, 0, "tarifa:jobs").Result()
			if err != nil || len(res) < 2 {
				log.Printf("[W%d] ⚠️  Error BRPOP: %v", id, err)
				continue
			}

			var job Job
			if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
				log.Printf("[W%d] ⚠️  Job inválido: %v", id, err)
				continue
			}

			start := time.Now()
			time.Sleep(1 * time.Millisecond)
			tarifa := rf.Predict(job.Consumo, job.Uso, job.Grupo, job.Empresa)
			elapsed := time.Since(start).Nanoseconds()

			key := "tarifa:result:" + job.ID
			if err := rdb.HSet(baseCtx, key, map[string]interface{}{
				"tarifa":      tarifa,
				"latency_ns":  elapsed,
				"worker_id":   workerName,
				"finished_at": time.Now().Unix(),
			}).Err(); err != nil {
				log.Printf("[W%d] ⚠️  Error guardando resultado: %v", id, err)
				continue
			}
			rdb.Expire(baseCtx, key, time.Hour)

			// --- 🆕 Actualización dinámica de estadísticas ---
			// --- Actualización dinámica de estadísticas ---
			tarifastats.IncrementTarifa(rdb, baseCtx, tarifa)
			tarifastats.IncrementWorker(rdb, baseCtx, workerName)
			tarifastats.IncrementTotals(rdb, baseCtx, elapsed)

			log.Printf("[%s] ✅ Job %s → tarifa=%d (%d ns)", workerName, job.ID, tarifa, elapsed)
		}
	}
}

// ------------ main: lanza el pool ------------
func main() {
	redisconn.Init()

	// Contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())

	// Nº de workers: default 4, se puede cambiar con WORKERS=#
	n := 4
	if env := os.Getenv("WORKERS"); env != "" {
		if v, err := strconv.Atoi(env); err == nil && v > 0 {
			n = v
		}
	}
	log.Printf("🚀 Iniciando pool con %d workers…", n)

	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go worker(i, ctx, &wg)
	}

	// Shutdown elegante con Ctrl-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("🛑 Señal recibida, cancelando workers…")
	cancel()                 // Cancelar el contexto
	redisconn.Client.Close() // Cerrar Redis
	wg.Wait()                // Esperar a que todos los workers terminen
}
