package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/smsohan/redis-autoscale/pkg/listlength"
)

var client *redis.Client
var redisConfig *listlength.Config

const defaultConsumtionTimeMils = time.Duration(100) * time.Microsecond

func main() {
	var err error
	redisConfig, client, err = listlength.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %q", err)
	}

	go consume()
	fmt.Println("Server listening on :8080")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func consume() {
	fmt.Println("Starting consumer")
	consumptionTime := defaultConsumtionTimeMils

	milsEnv := os.Getenv("REDIS_CONSUMPTION_TIME_MILS")
	if milsEnv != "" {
		mils, err := strconv.Atoi(milsEnv)
		if err != nil {
			log.Fatalf("Failed to start the consumer, invalid REDIS_CONSUMPTION_TIME_MILS: %q", err)
			return
		}
		consumptionTime = time.Duration(mils) * time.Microsecond
	}

	for {
		popped, err := client.LPop(redisConfig.ListName).Result()
		if err != nil {
			fmt.Print(".")
		} else {
			fmt.Printf("Consumed value:%q\n", popped)
		}
		time.Sleep(consumptionTime)
	}
}
