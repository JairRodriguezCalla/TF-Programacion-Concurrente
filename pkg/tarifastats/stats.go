package tarifastats

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// ---------- 1) Totales globales ---------------------------------------------

// IncrementTotals actualiza:
//   - total de jobs procesados
//   - suma de latencias (ns)
//   - contador de latencias (normalmente = total)
func IncrementTotals(rdb *redis.Client, ctx context.Context, latencyNs int64) {
	rdb.Incr(ctx, "tarifa:stats:total")
	rdb.IncrBy(ctx, "tarifa:stats:lat_total_ns", latencyNs)
	rdb.Incr(ctx, "tarifa:stats:lat_count")
}

// ---------- 2) Estadísticas por tarifa --------------------------------------

func IncrementTarifa(rdb *redis.Client, ctx context.Context, tarifa int) {
	rdb.Incr(ctx, "tarifa:stats:tarifa:"+strconv.Itoa(tarifa))
}

// ---------- 3) Estadísticas por worker --------------------------------------

func IncrementWorker(rdb *redis.Client, ctx context.Context, workerID string) {
	rdb.Incr(ctx, "tarifa:stats:worker:"+workerID)
}
