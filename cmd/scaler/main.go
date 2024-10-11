package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/smsohan/redis-autoscale/pkg/cloudrun"
	"github.com/smsohan/redis-autoscale/pkg/listlength"
)

var client *redis.Client
var redisConfig *listlength.Config

const MAX_INSTANCE_COUNT = 100

var consumerServiceFQN string

func main() {
	var err error
	redisConfig, client, err = listlength.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %q", err)
	}

	http.HandleFunc("/scale", scale)
	http.HandleFunc("/", length)

	consumerServiceFQN = fmt.Sprintf("projects/%s/locations/%s/services/%s",
		os.Getenv("CONSUMER_PROJECT_ID"), os.Getenv("CONSUMER_REGION"), os.Getenv("CONSUMER_SERVICE_NAME"))

	fmt.Println("Server listening on :8080")
	fmt.Println("== USAGE ==")
	fmt.Println("GET /length")
	fmt.Println("POST /scale")
	fmt.Println("== END USAGE ==")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func length(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		len := client.LLen(redisConfig.ListName)
		if len.Err() != nil {
			http.Error(w, fmt.Sprintf("Failed to get the length of %s: %q", redisConfig.ListName, len.Err()), http.StatusInternalServerError)
		}

		fmt.Fprintf(w, "List: %s, Length: %d, targetLength: %d\n", redisConfig.ListName, len.Val(), redisConfig.ListLength)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func scale(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		currentLength, err := client.LLen(redisConfig.ListName).Result()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get the length of %s: %q", redisConfig.ListName, err), http.StatusInternalServerError)
		}

		fmt.Printf("currentLenght: %d, targetLength: %d\n", currentLength, redisConfig.ListLength)

		currentInstanceCount, err := cloudrun.GetCurrentInstanceCount(consumerServiceFQN)
		if err != nil {
			fmt.Printf("Failed to get current instance count: %q\n", err)
		}

		maxCount, err := strconv.Atoi(os.Getenv("MAX_INSTANCE_COUNT"))
		if err != nil {
			http.Error(w, "error: invalid MAX_INSTANCE_COUNT", http.StatusInternalServerError)
		}

		targetInstanceCount := computeTargetInstanceCount(
			currentInstanceCount,
			maxCount,
			currentLength,
			redisConfig.ListLength,
		)

		if targetInstanceCount != currentInstanceCount {
			cloudrun.SetMinInstanceCount(consumerServiceFQN, targetInstanceCount)
		}
		fmt.Fprintf(w, "Listlength: %d, Instance count: %d -> %d\n", currentLength, currentInstanceCount, targetInstanceCount)

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func computeTargetInstanceCount(currentInstanceCount, maxInstanceCount int, currentListlength, targetListLength int64) int {
	// K8s HPA: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details
	usageRatio := float64(currentListlength) / float64(targetListLength)

	// scale up or down
	if currentInstanceCount != 0 {
		return int(math.Min(
			float64(maxInstanceCount),
			math.Ceil(float64(currentInstanceCount)*usageRatio),
		))
	}

	// possibly scale up to n
	return int(math.Ceil(usageRatio))
}
