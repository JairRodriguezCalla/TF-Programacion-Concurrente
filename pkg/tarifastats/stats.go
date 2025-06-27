package tarifastats

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Incrementa el contador acumulado para la tarifa predicha
func Increment(rdb *redis.Client, ctx context.Context, tarifa int) {
	key := "tarifa:stats"
	field := fmt.Sprintf("tarifa_%d", tarifa) // ej. tarifa_19
	rdb.HIncrBy(ctx, key, field, 1)
}
