REPO=us-central1-docker.pkg.dev/sohansm-project/cloud-run-source-deploy

PRODUCER_IMAGE=$REPO/producer
CGO_ENABLED=0 go build -o ./build ./cmd/producer
docker build -t $PRODUCER_IMAGE --build-arg BINARY=producer .
docker push $PRODUCER_IMAGE

gcloud run deploy redis-producer \
--set-env-vars REDIS_HOST=10.126.55.187,REDIS_PORT=6379,REDIS_LIST_NAME=mylist \
--region us-central1 \
--network=crf-vpc \
--subnet=crf-vpc \
--vpc-egress=private-ranges-only \
--no-allow-unauthenticated \
--image $PRODUCER_IMAGE

CONSUMER_IMAGE=$REPO/consumer
CGO_ENABLED=0 go build -o ./build ./cmd/consumer
docker build -t $CONSUMER_IMAGE --build-arg BINARY=consumer .
docker push $CONSUMER_IMAGE

gcloud alpha run deploy redis-consumer \
--set-env-vars MODE=CONSUMER,REDIS_HOST=10.126.55.187,REDIS_PORT=6379,REDIS_LIST_NAME=mylist,REDIS_CONSUMPTION_TIME_MILS=100 \
--region us-central1 \
--scaling manual \
--network=crf-vpc \
--subnet=crf-vpc \
--vpc-egress=private-ranges-only \
--no-allow-unauthenticated \
--image $CONSUMER_IMAGE

SCALER_IMAGE=$REPO/scaler
CGO_ENABLED=0 go build -o ./build ./cmd/scaler
docker build -t $SCALER_IMAGE --build-arg BINARY=scaler .
docker push $SCALER_IMAGE

gcloud alpha run deploy redis-scaler \
--set-env-vars REDIS_HOST=10.126.55.187,REDIS_PORT=6379,REDIS_LIST_NAME=mylist,REDIS_LIST_LENGTH=100,MAX_INSTANCE_COUNT=10,CONSUMER_PROJECT_ID=sohansm-project,CONSUMER_REGION=us-central1,CONSUMER_SERVICE_NAME=redis-consumer \
--region us-central1 \
--network=crf-vpc \
--subnet=crf-vpc \
--vpc-egress=private-ranges-only \
--no-allow-unauthenticated \
--image $SCALER_IMAGE