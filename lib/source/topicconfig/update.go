package topicconfig

import (
	"github.com/Shopify/sarama"
	"log"
)

func EnsureWithBroker(broker string, topic string, config map[string]string) (err error) {
	sconfig := sarama.NewConfig()
	sconfig.Version = sarama.V2_4_0_0
	admin, err := sarama.NewClusterAdmin([]string{broker}, sconfig)
	if err != nil {
		return err
	}

	temp := map[string]*string{}
	for key, value := range config {
		tempValue := value
		temp[key] = &tempValue
	}

	err = set(admin, topic, temp)
	if err != nil {
		log.Println("WARNING: ", err)
		err = create(admin, topic, temp)
	}

	return err
}

func set(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.AlterConfig(sarama.TopicResource, topic, config, false)
}

func create(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries:     config,
	}, false)
}
