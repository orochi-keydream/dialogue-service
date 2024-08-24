package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/orochi-keydream/dialogue-service/internal/config"
	"github.com/orochi-keydream/dialogue-service/internal/model"
	"github.com/orochi-keydream/dialogue-service/internal/service"
	"golang.org/x/net/context"
	"log"
	"sync"
)

func RunDialogueCommandConsumer(
	ctx context.Context,
	config config.KafkaConfig,
	c *DialogueCommandConsumer,
	wg *sync.WaitGroup,
) error {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cg, err := sarama.NewConsumerGroup(config.Brokers, "dialogue-service", cfg)

	if err != nil {
		panic(err)
	}

	topics := []string{config.Consumers.DialogueCommands.Topic}

	go func() {
		defer wg.Done()

		for {
			err = cg.Consume(ctx, topics, c)

			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}

				log.Panicln(err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

type DialogueCommandConsumer struct {
	appService *service.AppService
}

func NewDialogueCommandConsumer(appService *service.AppService) *DialogueCommandConsumer {
	return &DialogueCommandConsumer{appService: appService}
}

func (c *DialogueCommandConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *DialogueCommandConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *DialogueCommandConsumer) ConsumeClaim(cgs sarama.ConsumerGroupSession, cgc sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-cgc.Messages():
			if !ok {
				log.Println("Message channel was closed")
				return nil
			}

			log.Printf("Handling message with offset %v from %v topic\n", msg.Offset, msg.Topic)

			err := c.handle(cgs.Context(), msg.Value)

			if err != nil {
				log.Println(err)
				continue
			}

			cgs.MarkMessage(msg, "")
		case <-cgs.Context().Done():
			log.Println("ConsumeClaim: cancellation requested")
			return nil
		}
	}
}

func (c *DialogueCommandConsumer) handle(ctx context.Context, msg []byte) error {
	message := Message{}
	err := json.Unmarshal(msg, &message)

	if err != nil {
		return err
	}

	switch message.Command {
	case MessageCommandCommitMessage:
		return c.handleCommitMessage(ctx, message)
	case MessageCommandRollbackMessage:
		return c.handleRollbackMessage(ctx, message)
	default:
		return fmt.Errorf("unknown command: %v", message.Command)
	}

}

func (c *DialogueCommandConsumer) handleCommitMessage(ctx context.Context, msg Message) error {
	payload := CommitMessagePayload{}
	err := json.Unmarshal(msg.Payload, &payload)

	if err != nil {
		return err
	}

	cmd := model.CommitMessageCommand{
		CorrelationId: msg.CorrelationId,
		MessageId:     model.MessageId(payload.MessageId),
	}

	return c.appService.CommitMessage(ctx, cmd)
}

func (c *DialogueCommandConsumer) handleRollbackMessage(ctx context.Context, msg Message) error {
	payload := RollbackMessagePayload{}
	err := json.Unmarshal(msg.Payload, &payload)

	if err != nil {
		return err
	}

	cmd := model.RollbackMessageCommand{
		CorrelationId: msg.CorrelationId,
		MessageId:     model.MessageId(payload.MessageId),
	}

	return c.appService.RollbackMessage(ctx, cmd)
}

type MessageCommand string

const (
	MessageCommandCommitMessage   MessageCommand = "CommitMessage"
	MessageCommandRollbackMessage MessageCommand = "RollbackMessage"
)

type Message struct {
	CorrelationId string          `json:"correlationId"`
	Command       MessageCommand  `json:"command"`
	Payload       json.RawMessage `json:"payload"`
}

type CommitMessagePayload struct {
	MessageId int64 `json:"messageId"`
}

type RollbackMessagePayload struct {
	MessageId int64 `json:"messageId"`
}
