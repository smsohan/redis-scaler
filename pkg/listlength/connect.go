package listlength

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

func Connect() (*Config, *redis.Client, error) {
	redisConfig, err := ReadListConfigFromEnv()
	if err != nil {
		log.Fatalf("Error in reading config: %q\n", err)
	}
	fmt.Printf("Connecting to: %s\n", redisConfig.Address)
	client := redis.NewClient(&redis.Options{
		Addr: redisConfig.Address,
		DB:   redisConfig.DatabaseIndex,
	})

	pong := client.Ping()
	if pong.Err() != nil {
		fmt.Printf("Failed to connect: %q", pong.Err())
		return nil, nil, err
	}
	fmt.Println("Redis connected")
	return redisConfig, client, nil
}
