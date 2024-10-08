IMAGE=us-central1-docker.pkg.dev/sohansm-project/cloud-run-source-deploy/redis-scaler
CONSUMER_SERVICE=redis-consumer
gcloud builds submit --pack image=$IMAGE

gcloud run deploy redis-scaler \
--set-env-vars REDIS_HOST=10.126.55.187,REDIS_PORT=6379,REDIS_LIST_NAME=mylist,REDIS_LIST_LENGTH=100,MAX_INSTANCE_COUNT=10,CONSUMER_PROJECT_ID=sohansm-project,CONSUMER_REGION=us-central1,CONSUMER_SERVICE_NAME=$CONSUMER_SERVICE \
--region us-central1 \
--network=crf-vpc \
--subnet=crf-vpc \
--vpc-egress=private-ranges-only \
--no-allow-unauthenticated \
--image $IMAGE

gcloud alpha run deploy $CONSUMER_SERVICE \
--set-env-vars MODE=CONSUMER,REDIS_HOST=10.126.55.187,REDIS_PORT=6379,REDIS_LIST_NAME=mylist,REDIS_LIST_LENGTH=100,REDIS_CONSUMPTION_TIME_MILS=100 \
--region us-central1 \
--scaling manual \
--network=crf-vpc \
--subnet=crf-vpc \
--vpc-egress=private-ranges-only \
--no-allow-unauthenticated \
--image $IMAGE
