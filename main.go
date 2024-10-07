package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/smsohan/redis-autoscale/pkg/cloudrun"
	"github.com/smsohan/redis-autoscale/pkg/listlength"
)

var client *redis.Client
var redisConfig *listlength.Config

const DEFAULT_COUNT = 10
const DEFAULT_LIST_ITEM = "dummy"
const INSTANCE_COUNT_CACHE_KEY = "INSTANCE_COUNT"
const MAX_INSTANCE_COUNT = 100

func main() {
	redisConfig, client = connectToRedis()

	mode := os.Getenv("MODE")
	if mode == "CONSUMER" {
		go consume()
	} else {
		http.HandleFunc("/publish", publish)
		http.HandleFunc("/length", length)
		http.HandleFunc("/scale", scale)
		http.HandleFunc("/", home)

		fmt.Println("Server listening on :8080")
		fmt.Println("== USAGE ==")
		fmt.Println("POST /publish?count=10 # default 10")
		fmt.Println("GET /length")
		fmt.Println("POST /scale")
		fmt.Println("== END USAGE ==")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func connectToRedis() (*listlength.Config, *redis.Client) {
	redisConfig, err := listlength.ReadListConfigFromEnv()
	if err != nil {
		log.Fatalf("Error in reading config: %q\n", err)
	}
	fmt.Printf("Connecting to: %s\n", redisConfig.Address)
	client := redis.NewClient(&redis.Options{
		Addr: redisConfig.Address,
		DB:   0, // Use default DB
	})

	pong := client.Ping()
	if pong.Err() != nil {
		fmt.Printf("Failed to connect: %q", pong.Err())
	}
	fmt.Println("Redis connected")
	return redisConfig, client
}

func publish(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		query := r.URL.Query()

		item := query.Get("item")
		if item == "" {
			item = DEFAULT_LIST_ITEM
		}

		countStr := query.Get("count")
		count := DEFAULT_COUNT
		if countStr != "" {
			var err error
			count, err = strconv.Atoi(countStr)
			if err != nil {
				http.Error(w, "Invalid count parameter", http.StatusBadRequest)
				return
			}
		}

		for i := 0; i < count; i++ {
			err := client.RPush(redisConfig.ListName, item).Err()
			if err != nil {
				http.Error(w, "Failed to add item to list", http.StatusInternalServerError)
				return
			}
		}

		fmt.Fprintf(w, "%d items added to list: %s\n", count, item)
		log.Printf("%d items added to list: %s\n", count, item)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
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

func consume() {
	fmt.Println("Starting consumer")
	for {
		popped, err := client.LPop(redisConfig.ListName).Result()
		if err != nil {
			fmt.Print(".")
		} else {
			fmt.Printf("Consumed value:%q", popped)
		}

		count, err := getCurrentInstanceCount()
		if count <= 0 {
			count = 1
		}
		if err != nil {
			log.Fatalf("Error getting instance count: %q", err)
		}
		time.Sleep(redisConfig.ConsumptionTimeMils)
	}
}

func scale(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		currentLength, err := client.LLen(redisConfig.ListName).Result()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get the length of %s: %q", redisConfig.ListName, err), http.StatusInternalServerError)
		}

		fmt.Printf("currentLenght: %d, targetLength: %d\n", currentLength, redisConfig.ListLength)

		currentInstanceCount, err := getCurrentInstanceCount()
		if err != nil {
			fmt.Printf("Failed to get current instance count: %q\n", err)
		}

		var targetInstanceCount int
		maxCount, err := strconv.ParseFloat(os.Getenv("MAX_INSTANCE_COUNT"), 64)
		if err != nil {
			http.Error(w, "error: invalid MAX_INSTANCE_COUNT", http.StatusInternalServerError)
		}

		if currentLength > 0 {
			// K8s HPA: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details
			targetInstanceCount = int(math.Min(maxCount,
				float64(math.Ceil(float64(currentInstanceCount)*float64(currentLength)/float64(redisConfig.ListLength)))))

			if targetInstanceCount == 0 {
				targetInstanceCount = 1
			}
		}

		setInstanceCount(targetInstanceCount)
		fmt.Fprintf(w, "Listlength: %d, Instance count: %d -> %d", currentLength, currentInstanceCount, targetInstanceCount)

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getCurrentInstanceCount() (int, error) {
	return cloudrun.GetCurrentInstanceCount(os.Getenv("CONSUMER_PROJECT_ID"), os.Getenv("CONSUMER_REGION"), os.Getenv("CONSUMER_SERVICE_NAME"))
}

func setInstanceCount(count int) error {
	_, err := client.Set(INSTANCE_COUNT_CACHE_KEY, count, time.Duration(0)).Result()
	//TODO: Call Cloud Run API to adjust the instance count
	return err
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
}
