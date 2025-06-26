package redisconn

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	// Ctx se reutiliza en todo el proyecto
	Ctx = context.Background()
	// Client es el singleton de Redis
	Client *redis.Client
)

// Init crea (una sola vez) el cliente Redis para todo el proceso.
func Init() {
	// Si ya existe, no lo recreamos
	if Client != nil {
		return
	}

	// Variable de entorno opcional; si no existe usamos IPv4 local
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379" // ⚠️ evita localhost→IPv6 en Windows
	}

	Client = redis.NewClient(&redis.Options{
		Addr: addr,
		// Password y DB en blanco (usamos los defaults del contenedor)
	})
}
