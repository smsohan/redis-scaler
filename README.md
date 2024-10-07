# This is a HTTP application for simulating Redis list based auto-scaling in a containerized app

It has the following endpoints

```bash
POST /publish?count=<10> # publish 10 messages to a Redis list

POST /scale # triggers a scaling event, with a rudimentary scaling logic

GET /length # show the current length of the queue
```

It needs to run with a number of env variables as follows:

```bash

# Main web app
$ REDIS_HOST=localhost REDIS_PORT=6379 REDIS_LIST_NAME=mylist REDIS_LIST_LENGTH=100 CONSUMER_PROJECT_ID=sohansm-project CONSUMER_REGION=us-central1 CONSUMER_SERVICE_NAME=redis-consumer MAX_INSTANCE_COUNT=50 go run main.go

# Consumer
$ PORT=3000 REDIS_HOST=localhost REDIS_PORT=6379 REDIS_LIST_NAME=mylist MODE=CONSUMER REDIS_CONSUMPTION_TIME_MILS=100 go run main.go
```


