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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMaxDuration = 60 * time.Second
	envPrefix          = "KAFKA_"
	modeCreate         = "create"
	modeDelete         = "delete"
	modeListTopics     = "list-topics"
)

//nolint: gochecknoglobals
var (
	mode              = flag.String("mode", modeCreate, fmt.Sprintf("%s,%s,%s", modeCreate, modeDelete, modeListTopics))
	logLevel          = flag.String("log.level", "INFO", "log level")
	topics            = flag.String("topics", "", "topics name, separor comma")
	maxDur            = flag.Duration("duration", defaultMaxDuration, "wait for the operation to finish")
	numParts          = flag.Int("create.partition-count", -1, "partition count")
	replicationFactor = flag.Int("create.replication-factor", 1, "replication factor")
)

func main() { //nolint:funlen,cyclop
	flag.Parse()

	if *mode != modeCreate && *mode != modeDelete && *mode != modeListTopics {
		log.Fatalf("unknown mode, use -mode=%s or -mode=%s or -mode=%s",
			modeCreate,
			modeDelete,
			modeListTopics,
		)
	}

	if (*mode == modeCreate || *mode == modeDelete) && len(*topics) == 0 {
		log.Fatal("specify topic name, use -topics=topicName")
	}

	logLevelParsed, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(logLevelParsed)

	if log.GetLevel() >= log.DebugLevel {
		log.SetReportCaller(true)
	}

	configMap := kafka.ConfigMap{}

	for _, env := range os.Environ() {
		envName := strings.Split(env, "=")[0]
		name := strings.ToUpper(envName)

		if strings.HasPrefix(name, envPrefix) {
			// remove prefix
			name = strings.TrimPrefix(name, envPrefix)
			// make lovercase
			name = strings.ToLower(name)
			// replace _ with .
			name = strings.ReplaceAll(name, "_", ".")

			log.Debugf("env=%s,name=%s,value=%s", env, name, os.Getenv(envName))
			configMap[name] = os.Getenv(envName)
		}
	}

	vnum, vstr := kafka.LibraryVersion()
	log.Debugf("LibraryVersion: %s (0x%x)\n", vstr, vnum)
	log.Debugf("LinkInfo:       %s\n", kafka.LibrdkafkaLinkInfo)

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	adminClient, err := kafka.NewAdminClient(&configMap)
	if err != nil {
		log.Fatalf("Failed to create Admin client: %s\n", err)
	}
	defer adminClient.Close()

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var results []kafka.TopicResult

	switch *mode {
	case modeCreate:
		createTopics := make([]kafka.TopicSpecification, 0)
		for _, topic := range strings.Split(*topics, ",") {
			createTopics = append(createTopics, kafka.TopicSpecification{ //nolint:exhaustivestruct
				Topic:             topic,
				NumPartitions:     *numParts,
				ReplicationFactor: *replicationFactor,
			})
		}

		results, err = adminClient.CreateTopics(
			ctx,
			createTopics,
			kafka.SetAdminOperationTimeout(*maxDur))
	case modeDelete:
		results, err = adminClient.DeleteTopics(
			ctx,
			strings.Split(*topics, ","),
			kafka.SetAdminOperationTimeout(*maxDur))
	case modeListTopics:
		meta, err := adminClient.GetMetadata(nil, true, int(maxDur.Milliseconds()))
		if err != nil {
			log.Fatal(err) //nolint: gocritic
		}

		for _, topic := range meta.Topics {
			fmt.Printf("%s\n", topic.Topic) //nolint: forbidigo
		}

	default:
		log.Fatalf("unknown mode %s", *mode)
	}

	if err != nil {
		log.Fatalf("Failed to %s topic: %v\n", *mode, err)
	}

	for _, result := range results {
		log.Infof("mode=%s, %s\n", *mode, result)
	}
}
