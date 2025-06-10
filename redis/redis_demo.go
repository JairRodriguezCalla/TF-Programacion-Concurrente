package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	err := client.Set(ctx, "prediccion:001", "BT5B", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "prediccion:001").Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Resultado recuperado desde Redis: %s\n", val)
}
