package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/orochi-keydream/dialogue-service/internal/model"
)

type Repository interface {
	AddMessage(ctx context.Context, msg *model.Message, tx *sql.Tx) (model.MessageId, error)
	GetMessages(ctx context.Context, chatId model.ChatId, tx *sql.Tx) ([]*model.Message, error)
}

type AppService struct {
	repository Repository
}

func NewAppService(repository Repository) *AppService {
	return &AppService{
		repository: repository,
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
	}

	messageId, err := s.repository.AddMessage(ctx, msg, nil)

	if err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("Message %v sent to chat %v", messageId, chatId))

	return nil
}

func (s *AppService) GetMessages(ctx context.Context, cmd model.GetMessagesCommand) ([]*model.Message, error) {
	chatId := s.buildChatId(cmd.FromUserId, cmd.ToUserId)
	messages, err := s.repository.GetMessages(ctx, chatId, nil)

	slog.InfoContext(ctx, fmt.Sprintf("Got %v messages from chat %v", len(messages), chatId))

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *AppService) buildChatId(firstUser, secondUser model.UserId) model.ChatId {
	if firstUser > secondUser {
		return model.ChatId(fmt.Sprintf("%s_%s", secondUser, firstUser))
	} else {
		return model.ChatId(fmt.Sprintf("%s_%s", firstUser, secondUser))
	}
}
