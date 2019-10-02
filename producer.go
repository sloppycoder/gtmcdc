package gtmcdc

import (
	"errors"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

var cdcProducer sarama.SyncProducer
var cdcTopic string

func CleanupProducer() {
	if cdcProducer != nil {
		log.Debug("cleanup producer")
		_ = cdcProducer.Close()
	}
}

func PublishMessage(message string) error {
	if cdcProducer == nil {
		return errors.New("producer not available")
	}

	_, _, err := cdcProducer.SendMessage(&sarama.ProducerMessage{
		Topic: cdcTopic,
		Value: sarama.StringEncoder(message),
	})

	if err != nil {
		log.Info("Unable to publish message")
		return err
	}

	return nil
}

func InitProducer(brokers, topic string) error {
	brokerList := strings.Split(brokers, ",")
	if len(brokerList) <= 1 || brokerList[0] == "off" || topic == "" {
		return errors.New("invalid kafka broker list specified")
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Errorln("Failed to start Sarama producer:", err)
		return err
	}

	SetProducer(producer)

	return nil
}

func SetProducer(producer sarama.SyncProducer) {
	cdcProducer = producer
}

func IsKafkaAvailable() bool {
	return (cdcProducer != nil)
}
