package listlength

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Address       string
	Password      string
	ListName      string
	ListLength    int64
	DatabaseIndex int
}

const defaultListLength = 10

func ReadListConfigFromEnv() (*Config, error) {

	length, err := readIntFromEnv("REDIS_LIST_LENGTH", defaultListLength)
	if err != nil {
		return nil, err
	}

	index, err := readIntFromEnv("REDIS_DATABASE_INDEX", 0)
	if err != nil {
		return nil, err
	}

	var config = &Config{
		Address:       fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:      os.Getenv("REDIS_PASSWORD"),
		ListName:      os.Getenv("REDIS_LIST_NAME"),
		ListLength:    length,
		DatabaseIndex: int(index),
	}

	return config, nil
}

func readIntFromEnv(key string, fallback int64) (int64, error) {
	val := fallback
	var err error
	env := os.Getenv(key)
	if env != "" {
		val, err = strconv.ParseInt(env, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return val, nil
}
