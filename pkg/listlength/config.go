package listlength

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address             string
	Password            string
	ListName            string
	ListLength          int64
	EnableTLS           string
	DatabaseIndex       string
	ConsumptionTimeMils time.Duration
}

const DEFUALT_CONSUMPTION_MILS = 100
const DEFUALT_LIST_LENGTH = 10

func ReadListConfigFromEnv() (*Config, error) {

	length, err := readIntFromEnv("REDIS_LIST_LENGTH", DEFUALT_LIST_LENGTH)
	if err != nil {
		return nil, err
	}

	mils, err := readIntFromEnv("REDIS_CONSUMPTION_TIME_MILS", DEFUALT_CONSUMPTION_MILS)
	if err != nil {
		return nil, err
	}

	var config = &Config{
		Address:             fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:            os.Getenv("REDIS_PASSWORD"),
		ListName:            os.Getenv("REDIS_LIST_NAME"),
		ListLength:          length,
		EnableTLS:           os.Getenv("REDIS_ENABLE_TLS"),
		DatabaseIndex:       os.Getenv("REDIS_DATABASE_INDEX"),
		ConsumptionTimeMils: time.Duration(mils) * time.Millisecond,
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
