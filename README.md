# This is a HTTP application for simulating Redis list based auto-scaling in a containerized app

It needs to run with a number of env variables as follows:

```bash
# Producer
$ REDIS_HOST=localhost REDIS_PORT=6379 REDIS_LIST_NAME=mylist go run ./cmd/producer
$ curl <app url>?count=100 -XPOST # publishes a 100 items to the list

# Scaler
$ PORT=8081 REDIS_HOST=localhost REDIS_PORT=6379 REDIS_LIST_NAME=mylist REDIS_LIST_LENGTH=100 CONSUMER_PROJECT_ID=sohansm-project CONSUMER_REGION=us-central1 CONSUMER_SERVICE_NAME=redis-consumer MAX_INSTANCE_COUNT=50 go run ./cmd/scaler
$ curl <app url> # returns the current size of the list
$ curl <app url>/scale -XPOST # scales the consumer based on the following algorithm

# Consumer
$ PORT=8082 REDIS_HOST=localhost REDIS_PORT=6379 REDIS_LIST_NAME=mylist REDIS_CONSUMPTION_TIME_MILS=100 go run ./cmd/consumer
```

It's a single app, but serves 3 purposes:

1. Publish: allows publishing entries to a Redis list to simulate a publisher
2. Consume: a consumer that pops messages from the list, using sleep to simulate processing time
3. Scale: looks at the Redis list and adjusts the number of consumer instances using the algorithm as described below

## Scaling Algorithm

The scaler uses a naive scaling algorithm, borrowed from [Kubernetes HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details) that works as follows:

$$
targetInstanceCount = \begin{cases}
  ceil (\frac{currentListLength}{targetListLength} ) & if \ currentInstanceCount == 0 \\
  ceil (currentInstanceCount * \frac{currentListLength}{targetListLength} ) & \text otherwise
\end{cases}
$$

This implementation is inspired by the [Redis List Scaler in Keda](https://keda.sh/docs/1.4/scalers/redis-lists/).