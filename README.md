# Kafka topics administration
Since librdkafka v1.6.0 - the consumer no longer triggers topic auto creation (regardless of allow.auto.create.topics=..) for subscribed topics. The recommendation is not to rely on auto topic creation at all, but to use the Admin API to create topics.

This tool can create/delete/list kafka topics with Admin API

# Usage
```bash

cat <<EOF >>.env
KAFKA_BOOTSTRAP_SERVERS=broker:9092
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISMS=PLAIN
KAFKA_SASL_USERNAME=login
KAFKA_SASL_PASSWORD=password
EOF

# use latest docker image
docker pull paskalmaksim/kafka-topic-admin:latest

# create topics
docker run -it --rm \
  --env-file=.env \
  paskalmaksim/kafka-topic-admin:latest \
  -mode=create \
  -topics=test1,test2

# delete topics
docker run -it --rm \
  --env-file=.env \
  paskalmaksim/kafka-topic-admin:latest \
  -mode=delete \
  -topics=test1,test2

# list all topics
docker run -it --rm \
  --env-file=.env \
  paskalmaksim/kafka-topic-admin:latest \
  -mode=list-topics
```