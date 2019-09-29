package gtmcdc

import (
	"errors"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

var cdcProducer sarama.SyncProducer

func CleanupProducer() {
	if cdcProducer != nil {
		log.Debug("cleanup producer")
		_ = cdcProducer.Close()
	}
}

func PublishMessage(topic, message string) error {
	if cdcProducer == nil {
		return errors.New("producer not inialized")
	}

	_, _, err := cdcProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	})

	if err != nil {
		log.Info("Unable to publish message")
		return err
	}

	return nil
}

func NewCDCProducer(brokerList []string) {
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
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	cdcProducer = producer
}
