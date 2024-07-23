package repository

import (
	"context"
	"database/sql"

	"github.com/orochi-keydream/dialogue-service/internal/model"
)

type DialogRepository struct {
	conn *sql.DB
}

func NewDialogRepository(conn *sql.DB) *DialogRepository {
	return &DialogRepository{
		conn: conn,
	}
}

func (r *DialogRepository) AddMessage(
	ctx context.Context,
	msg *model.Message,
	tx *sql.Tx,
) (model.MessageId, error) {
	const query = `
		insert into messages
		(
			chat_id,
			sent_at,
			from_user_id,
			to_user_id,
			text
		)
		values ($1, $2, $3, $4, $5)
		returning message_id`

	var ec IExecutionContext

	if tx == nil {
		ec = r.conn
	} else {
		ec = tx
	}

	row := ec.QueryRowContext(
		ctx,
		query,
		msg.ChatId,
		msg.SentAt,
		msg.FromUserId,
		msg.ToUserId,
		msg.Text)

	if row.Err() != nil {
		return 0, row.Err()
	}

	var messageId model.MessageId
	err := row.Scan(&messageId)

	if err != nil {
		return 0, err
	}

	return messageId, nil
}

func (r *DialogRepository) GetMessages(
	ctx context.Context,
	chatId model.ChatId,
	tx *sql.Tx,
) ([]*model.Message, error) {
	const query = `
		select
			message_id,
			chat_id,
			sent_at,
			from_user_id,
			to_user_id,
			text
		from messages
		where chat_id = $1
		order by sent_at desc
		`

	var ec IExecutionContext

	if tx == nil {
		ec = r.conn
	} else {
		ec = tx
	}

	rows, err := ec.QueryContext(ctx, query, chatId)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	messages := []*model.Message{}

	for rows.Next() {
		var msg model.Message

		err = rows.Scan(
			&msg.MessageId,
			&msg.ChatId,
			&msg.SentAt,
			&msg.FromUserId,
			&msg.ToUserId,
			&msg.Text)

		if err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}
