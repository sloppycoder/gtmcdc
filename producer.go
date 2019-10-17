package gtmcdc

import (
	"errors"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
}

func (p *Producer) CleanupProducer() {
	if p != nil && p.syncProducer != nil {
		log.Debug("cleanup producer")
		_ = p.syncProducer.Close()
	}
}

func (p *Producer) PublishMessage(message string) error {
	if p.syncProducer == nil {
		return errors.New("producer not available")
	}

	_, _, err := p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(message),
	})

	if err != nil {
		log.Info("Unable to publish message")
		return err
	}

	return nil
}

func InitProducer(brokers, topic string) (*Producer, error) {
	brokerList := strings.Split(brokers, ",")
	if len(brokerList) < 1 || brokerList[0] == "off" || topic == "" {
		return nil, errors.New("invalid kafka broker list or topic specified")
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Version = sarama.MaxVersion

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.
	syncProducer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Errorln("Failed to start Sarama producer:", err)
		return nil, err
	}

	producer := &Producer{
		syncProducer: syncProducer,
		topic:        topic,
	}

	return producer, nil
}

func (p *Producer) IsKafkaAvailable() bool {
	return p != nil && p.syncProducer != nil
}
