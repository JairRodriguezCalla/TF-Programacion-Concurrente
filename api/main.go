package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

// -------- helpers ----------
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// -------- handlers ----------
func status(w http.ResponseWriter, r *http.Request) {
	total, _ := rdb.Get(ctx, "tarifa:stats:total").Int64()
	sumLat, _ := rdb.Get(ctx, "tarifa:stats:lat_total_ns").Int64()
	latCnt, _ := rdb.Get(ctx, "tarifa:stats:lat_count").Int64()

	var prom int64
	if latCnt > 0 {
		prom = sumLat / latCnt
	}

	respond(w, map[string]any{
		"total_jobs":      total,
		"latency_ns_prom": prom,
	})
}

func statsTarifas(w http.ResponseWriter, r *http.Request) {
	keys, _ := rdb.Keys(ctx, "tarifa:stats:tarifa:*").Result()
	data := map[string]int64{}
	for _, k := range keys {
		v, _ := rdb.Get(ctx, k).Int64()
		t := k[len("tarifa:stats:tarifa:"):]
		data[t] = v
	}
	respond(w, data)
}

func statsWorkers(w http.ResponseWriter, r *http.Request) {
	keys, _ := rdb.Keys(ctx, "tarifa:stats:worker:*").Result()
	data := map[string]int64{}
	for _, k := range keys {
		v, _ := rdb.Get(ctx, k).Int64()
		id := k[len("tarifa:stats:worker:"):]
		data[id] = v
	}
	respond(w, data)
}

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	rdb = redis.NewClient(&redis.Options{Addr: addr})

	http.HandleFunc("/status", status)
	http.HandleFunc("/stats/tarifas", statsTarifas)
	http.HandleFunc("/stats/workers", statsWorkers)

	port := getenvDefault("PORT", "8080")
	log.Printf("🌐 API escuchando en :%s ...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
