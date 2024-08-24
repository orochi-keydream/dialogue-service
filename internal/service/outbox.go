package service

import (
	"context"
	"database/sql"
	"github.com/orochi-keydream/dialogue-service/internal/model"
)

// TODO: Move it somewhere else.
type ITransactionManager interface {
	Begin(ctx context.Context) (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error
}

type OutboxService struct {
	outboxRepository   IOutboxRepository
	producer           IOutboxProducer
	transactionManager ITransactionManager
}

type IOutboxProducer interface {
	SendMessage(message *model.OutboxMessage) error
}

func NewOutboxService(
	repository IOutboxRepository,
	producer IOutboxProducer,
	transactionManager ITransactionManager,
) *OutboxService {
	return &OutboxService{
		outboxRepository:   repository,
		producer:           producer,
		transactionManager: transactionManager,
	}
}

func (s *OutboxService) Send(ctx context.Context) error {
	messages, err := s.outboxRepository.GetUnsent(ctx, nil)

	if err != nil {
		return err
	}

	for _, message := range messages {
		err = s.producer.SendMessage(message)

		if err != nil {
			return err
		}

		message.IsSent = true
	}

	return s.outboxRepository.Update(ctx, messages, nil)
}
