package main

import (
	"encoding/json"
	"log"
	"time"

	"tf/nodes/pkg/redisconn" // ✅ usa ruta relativa dentro de módulo nodes
	"tf/nodes/rf"            // ✅ tu modelo Random Forest
)

type Job struct {
	ID      string  `json:"id"`
	Consumo float64 `json:"consumo"`
	Uso     int     `json:"uso"`
	Grupo   int     `json:"grupo"`
	Empresa int     `json:"empresa"`
}

func main() {
	redisconn.Init()
	rdb := redisconn.Client
	ctx := redisconn.Ctx

	log.Println("🚀 Worker iniciado — esperando trabajos…")

	for {
		res, err := rdb.BRPop(ctx, 0, "tarifa:jobs").Result()
		if err != nil || len(res) < 2 {
			log.Printf("⚠️  Error BRPOP: %v", err)
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			log.Printf("⚠️  Job inválido: %v", err)
			continue
		}

		start := time.Now()
		tarifa := rf.Predict(job.Consumo, job.Uso, job.Grupo, job.Empresa)
		time.Sleep(1 * time.Millisecond) //Comprobar funcionamiento
		elapsed := time.Since(start)

		key := "tarifa:result:" + job.ID
		err = rdb.HSet(ctx, key, map[string]interface{}{
			"tarifa":   tarifa,
			"latency":  elapsed.Seconds(),
			"finished": time.Now().Unix(),
		}).Err()
		if err != nil {
			log.Printf("⚠️  Error guardando resultado: %v", err)
			continue
		}

		rdb.Expire(ctx, key, time.Hour)

		log.Printf("✅ Job %s → tarifa=%d (%d ns)", job.ID, tarifa, elapsed.Nanoseconds())

	}
}
