package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/smsohan/redis-autoscale/pkg/listlength"
)

var client *redis.Client
var redisConfig *listlength.Config

const DEFAULT_COUNT = 10
const DEFAULT_LIST_ITEM = "dummy"

func main() {
	var err error
	redisConfig, client, err = listlength.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %q", err)
	}

	http.HandleFunc("/", publish)

	fmt.Println("Server listening on :8080")
	fmt.Println("== USAGE ==")
	fmt.Println("POST /?count=10 # default 10")
	fmt.Println("== END USAGE ==")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
