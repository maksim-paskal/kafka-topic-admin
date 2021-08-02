/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultMaxDuration = 60 * time.Second
)

//nolint: gochecknoglobals
var (
	mode              = flag.String("mode", "create", "create or delete")
	bootstrap         = flag.String("bootstrap", "bootstrap.yaml", "config yaml")
	topics            = flag.String("topics", "", "topics name, separor comma")
	maxDur            = flag.Duration("duration", defaultMaxDuration, "wait for the operation to finish")
	numParts          = flag.Int("create.partition-count", -1, "partition count")
	replicationFactor = flag.Int("create.replication-factor", -1, "replication factor")
)

func main() { //nolint:funlen,cyclop
	flag.Parse()

	if *mode != "create" && *mode != "delete" {
		log.Fatal("unknown node, use -mode=create or -mode=delete")
	}

	if len(*topics) == 0 {
		log.Fatal("specify topic name, use -topics=topicName")
	}

	configMap := kafka.ConfigMap{}

	config, err := ioutil.ReadFile(*bootstrap)
	if err != nil {
		log.Fatalf("Failed to load config: %s\n", err)
	}

	err = yaml.Unmarshal(config, &configMap)
	if err != nil {
		log.Fatalf("Failed to load parse yaml: %s\n", err)
	}

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&configMap)
	if err != nil {
		log.Fatalf("Failed to create Admin client: %s\n", err)
	}
	defer a.Close()

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var results []kafka.TopicResult

	switch *mode {
	case "create":
		createTopics := make([]kafka.TopicSpecification, 0)
		for _, topic := range strings.Split(*topics, ",") {
			createTopics = append(createTopics, kafka.TopicSpecification{ //nolint:exhaustivestruct
				Topic:             topic,
				NumPartitions:     *numParts,
				ReplicationFactor: *replicationFactor,
			})
		}

		results, err = a.CreateTopics(
			ctx,
			createTopics,
			kafka.SetAdminOperationTimeout(*maxDur))
	case "delete":
		results, err = a.DeleteTopics(
			ctx,
			strings.Split(*topics, ","),
			kafka.SetAdminOperationTimeout(*maxDur))
	default:
		log.Fatal("unknown node") //nolint:gocritic
	}

	if err != nil {
		log.Fatalf("Failed to %s topic: %v\n", *mode, err)
	}

	for _, result := range results {
		log.Infof("mode=%s, %s\n", *mode, result)
	}
}
