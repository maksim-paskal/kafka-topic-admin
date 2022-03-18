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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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
	topicFile         = flag.String("topics-file", "", "topics yaml")
	maxDur            = flag.Duration("duration", defaultMaxDuration, "wait for the operation to finish")
	numParts          = flag.Int("create.partition-count", -1, "partition count")
	replicationFactor = flag.Int("create.replication-factor", -1, "replication factor")
)

type kafkaTopicsFile struct {
	Topics []string `yaml:"topics"`
}

func main() { //nolint:funlen,cyclop,gocognit
	flag.Parse()

	logLevelParsed, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithError(err).Fatal("can not parse level")
	}

	log.SetLevel(logLevelParsed)

	if log.GetLevel() >= log.DebugLevel {
		log.SetReportCaller(true)
	}

	if *mode != modeCreate && *mode != modeDelete && *mode != modeListTopics {
		log.Fatalf("unknown mode, use -mode=%s or -mode=%s or -mode=%s",
			modeCreate,
			modeDelete,
			modeListTopics,
		)
	}

	// all topics
	inputTopics := make([]string, 0)

	if len(*topicFile) > 0 {
		topicsBytes, err := ioutil.ReadFile(*topicFile)
		if err != nil {
			log.WithError(err).Fatalf("can not read yaml file")
		}

		topicsFile := kafkaTopicsFile{}

		err = yaml.Unmarshal(topicsBytes, &topicsFile)
		if err != nil {
			log.WithError(err).Fatalf("can not parse yaml")
		}

		inputTopics = topicsFile.Topics
	}

	if len(*topics) > 0 {
		inputTopics = strings.Split(*topics, ",")
	}

	if (*mode == modeCreate || *mode == modeDelete) && len(inputTopics) == 0 {
		log.Fatal("specify topics name, use -topics=topicName or -topics-file=kafka-topics.yaml")
	}

	log.Debug(inputTopics)

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
		log.WithError(err).Fatal("Failed to create Admin client")
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

		for _, topic := range inputTopics {
			kafkaSpec := kafka.TopicSpecification{
				Topic:         topic,
				NumPartitions: *numParts,
			}

			if *replicationFactor > 0 {
				kafkaSpec.ReplicationFactor = *replicationFactor
			}

			createTopics = append(createTopics, kafkaSpec)
		}

		results, err = adminClient.CreateTopics(
			ctx,
			createTopics,
			kafka.SetAdminOperationTimeout(*maxDur))
	case modeDelete:
		results, err = adminClient.DeleteTopics(
			ctx,
			inputTopics,
			kafka.SetAdminOperationTimeout(*maxDur))
	case modeListTopics:
		meta, err := adminClient.GetMetadata(nil, true, int(maxDur.Milliseconds()))
		if err != nil {
			log.WithError(err).Fatal("can not get metadata")
		}

		for _, topic := range meta.Topics {
			fmt.Printf("%s\n", topic.Topic) //nolint: forbidigo
		}

	default:
		log.Fatalf("unknown mode %s", *mode) //nolint:gocritic
	}

	if err != nil {
		log.WithError(err).Fatalf("Failed to %s topic", *mode)
	}

	isFatal := false

	for _, result := range results {
		log := log.WithField("mode", *mode)

		switch result.Error.Code() {
		case kafka.ErrNoError:
			log.Info(result)
		case kafka.ErrTopicAlreadyExists:
			log.Warn(result)
		default:
			isFatal = true

			log.WithError(result.Error).Error(result)
		}
	}

	if isFatal {
		os.Exit(-1)
	}
}
