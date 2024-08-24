package producer

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/orochi-keydream/dialogue-service/internal/config"
	"github.com/orochi-keydream/dialogue-service/internal/model"
)

type CounterCommandProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewCounterCommandProducer(config config.KafkaConfig) (*CounterCommandProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(config.Brokers, cfg)

	if err != nil {
		return nil, err
	}

	p := &CounterCommandProducer{
		producer: producer,
		topic:    config.Producers.CounterCommands.Topic,
	}

	return p, nil
}

func (p *CounterCommandProducer) SendMessage(message *model.OutboxMessage) error {
	var (
		messageKeyBytes   []byte
		messageValueBytes []byte
	)

	switch message.Type {
	case model.OutboxMessageTypeAddNewUnreadMessage:
		messageKey := message.MessageKey.(string)
		messageValue := message.MessageValue.(model.AddNewUnreadMessage)

		messageKeyBytes = []byte(messageKey)

		bytes, err := mapAddNewUnreadMessageToBytes(messageValue)

		if err != nil {
			return err
		}

		messageValueBytes = bytes
	default:
		return fmt.Errorf("Unsupported message type: %s", message.Type)
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(messageKeyBytes),
		Value: sarama.StringEncoder(messageValueBytes),
		Topic: p.topic,
	}

	_, _, err := p.producer.SendMessage(msg)

	return err
}

func mapAddNewUnreadMessageToBytes(message model.AddNewUnreadMessage) ([]byte, error) {
	dto := struct {
		CorrelationId string `json:"correlationId"`
		UserId        string `json:"userId"`
		ChatId        string `json:"chatId"`
		MessageId     int64  `json:"messageId"`
	}{
		CorrelationId: message.CorrelationId,
		UserId:        string(message.UserId),
		ChatId:        string(message.ChatId),
		MessageId:     int64(message.MessageId),
	}

	return json.Marshal(dto)
}
