package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"

	"github.com/orochi-keydream/dialogue-service/internal/model"
)

type IDialogueRepository interface {
	AddMessage(ctx context.Context, msg *model.Message, tx *sql.Tx) (model.MessageId, error)
	GetSentMessages(ctx context.Context, chatId model.ChatId, tx *sql.Tx) ([]*model.Message, error)
	GetMessage(ctx context.Context, id model.MessageId, tx *sql.Tx) (*model.Message, error)
	UpdateMessage(ctx context.Context, msg *model.Message, tx *sql.Tx) error
}

type IOutboxRepository interface {
	Add(ctx context.Context, message *model.OutboxMessage, tx *sql.Tx) error
	GetUnsent(ctx context.Context, tx *sql.Tx) ([]*model.OutboxMessage, error)
	Update(ctx context.Context, messages []*model.OutboxMessage, tx *sql.Tx) error
}

type ICommandRepository interface {
	Add(ctx context.Context, correlationId string, tx *sql.Tx) error
	Exists(ctx context.Context, correlationId string, tx *sql.Tx) (bool, error)
}

type AppService struct {
	dialogueRepository IDialogueRepository
	outboxRepository   IOutboxRepository
	commandRepository  ICommandRepository
	transactionManager ITransactionManager
}

func NewAppService(
	dialogueRepository IDialogueRepository,
	outboxRepository IOutboxRepository,
	commandRepository ICommandRepository,
	transactionManager ITransactionManager,
) *AppService {
	return &AppService{
		dialogueRepository: dialogueRepository,
		outboxRepository:   outboxRepository,
		commandRepository:  commandRepository,
		transactionManager: transactionManager,
	}
}

func (s *AppService) SendMessage(ctx context.Context, cmd model.SendMessageCommand) error {
	chatId := s.buildChatId(cmd.FromUserId, cmd.ToUserId)

	msg := &model.Message{
		ChatId:     chatId,
		FromUserId: cmd.FromUserId,
		ToUserId:   cmd.ToUserId,
		Text:       cmd.Text,
		SentAt:     time.Now().UTC(),
		State:      model.MessageStatePending,
	}

	tx, err := s.transactionManager.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	messageId, err := s.dialogueRepository.AddMessage(ctx, msg, tx)

	if err != nil {
		return err
	}

	messageValue := model.AddNewUnreadMessage{
		CorrelationId: uuid.New().String(),
		UserId:        msg.ToUserId,
		ChatId:        msg.ChatId,
		MessageId:     messageId,
	}

	outboxMessage := &model.OutboxMessage{
		Type:         model.OutboxMessageTypeAddNewUnreadMessage,
		MessageKey:   msg.ChatId,
		MessageValue: messageValue,
		IsSent:       false,
	}

	err = s.outboxRepository.Add(ctx, outboxMessage, tx)

	if err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("Message %v sent to chat %v", messageId, chatId))

	tx.Commit()

	return nil
}

func (s *AppService) GetMessages(ctx context.Context, cmd model.GetMessagesCommand) ([]*model.Message, error) {
	chatId := s.buildChatId(cmd.FromUserId, cmd.ToUserId)

	messages, err := s.dialogueRepository.GetSentMessages(ctx, chatId, nil)

	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, fmt.Sprintf("Got %v messages from chat %v", len(messages), chatId))

	return messages, nil
}

func (s *AppService) CommitMessage(ctx context.Context, cmd model.CommitMessageCommand) error {
	exists, err := s.commandRepository.Exists(ctx, cmd.CorrelationId, nil)

	if err != nil {
		return err
	}

	if exists {
		slog.Info(fmt.Sprintf("Command with correlation ID %v was handled before", cmd.CorrelationId))
		return nil
	}

	message, err := s.dialogueRepository.GetMessage(ctx, cmd.MessageId, nil)

	if err != nil {
		return err
	}

	message.State = model.MessageStateSent

	tx, err := s.transactionManager.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = s.dialogueRepository.UpdateMessage(ctx, message, tx)

	if err != nil {
		return err
	}

	err = s.commandRepository.Add(ctx, cmd.CorrelationId, tx)

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("Message %v has been committed", cmd.MessageId))

	return nil
}

func (s *AppService) RollbackMessage(ctx context.Context, cmd model.RollbackMessageCommand) error {
	exists, err := s.commandRepository.Exists(ctx, cmd.CorrelationId, nil)

	if err != nil {
		return err
	}

	if exists {
		slog.Info(fmt.Sprintf("Command with correlation ID %v was handled before", cmd.CorrelationId))
		return nil
	}

	message, err := s.dialogueRepository.GetMessage(ctx, cmd.MessageId, nil)

	if err != nil {
		return err
	}

	message.State = model.MessageStateRemoved

	tx, err := s.transactionManager.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = s.dialogueRepository.UpdateMessage(ctx, message, tx)

	if err != nil {
		return err
	}

	err = s.commandRepository.Add(ctx, cmd.CorrelationId, tx)

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Message %v has been removed due to rollback", message.MessageId))

	return nil
}

func (s *AppService) buildChatId(firstUser, secondUser model.UserId) model.ChatId {
	if firstUser > secondUser {
		return model.ChatId(fmt.Sprintf("%s_%s", secondUser, firstUser))
	} else {
		return model.ChatId(fmt.Sprintf("%s_%s", firstUser, secondUser))
	}
}
