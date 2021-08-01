# Kafka topics administration
Since librdkafka v1.6.0 - the consumer no longer triggers topic auto creation (regardless of allow.auto.create.topics=..) for subscribed topics. The recommendation is not to rely on auto topic creation at all, but to use the Admin API to create topics.

This tool can create/delete kafka topics with Admin API

# Usage
```bash
# create bootstrap.yaml with kafka auth
cat <<EOF >>bootstrap.yaml
bootstrap.servers: broker:9092
security.protocol: SASL_SSL
sasl.mechanisms: PLAIN
sasl.username: login
sasl.password: password
EOF

# use latest docker image
docker pull paskalmaksim/kafka-topic-admin:latest

# create topics
docker run -it --rm \
  -v $(pwd)/bootstrap.yaml:/app/bootstrap.yaml \
  paskalmaksim/kafka-topic-admin:latest \
  /app/kafka-topic-creation \
  -mode=create \
  -bootstrap=/app/bootstrap.yaml \
  -topics=test1,test2

# delete topics
docker run -it --rm \
  -v $(pwd)/bootstrap.yaml:/app/bootstrap.yaml \
  paskalmaksim/kafka-topic-admin:latest \
  /app/kafka-topic-creation \
  -mode=delete \
  -bootstrap=/app/bootstrap.yaml \
  -topics=test1,test2
```