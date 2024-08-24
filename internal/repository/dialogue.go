package repository

import (
	"context"
	"database/sql"

	"github.com/orochi-keydream/dialogue-service/internal/model"
)

type DialogRepository struct {
	db *sql.DB
}

func NewDialogueRepository(conn *sql.DB) *DialogRepository {
	return &DialogRepository{
		db: conn,
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
			text,
			state
		)
		values ($1, $2, $3, $4, $5, $6)
		returning message_id`

	var ec IExecutionContext

	if tx == nil {
		ec = r.db
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
		msg.Text,
		msg.State)

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

func (r *DialogRepository) GetSentMessages(
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
			text,
			state
		from messages
		where
			chat_id = $1 and
			state = $2
		order by sent_at desc
		`

	var ec IExecutionContext

	if tx == nil {
		ec = r.db
	} else {
		ec = tx
	}

	state := model.MessageStateSent

	rows, err := ec.QueryContext(ctx, query, chatId, state)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var messages []*model.Message

	for rows.Next() {
		var msg model.Message

		err = rows.Scan(
			&msg.MessageId,
			&msg.ChatId,
			&msg.SentAt,
			&msg.FromUserId,
			&msg.ToUserId,
			&msg.Text,
			&msg.State)

		if err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *DialogRepository) GetMessage(ctx context.Context, id model.MessageId, tx *sql.Tx) (*model.Message, error) {
	const query = `
		select
			message_id,
			chat_id,
			sent_at,
			from_user_id,
			to_user_id,
			text,
			state
		from messages
		where message_id = $1`

	var ec IExecutionContext

	if tx == nil {
		ec = r.db
	} else {
		ec = tx
	}

	row := ec.QueryRowContext(ctx, query, id)

	if row.Err() != nil {
		return nil, row.Err()
	}

	message := &model.Message{}

	err := row.Scan(
		&message.MessageId,
		&message.ChatId,
		&message.SentAt,
		&message.FromUserId,
		&message.ToUserId,
		&message.Text,
		&message.State,
	)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *DialogRepository) UpdateMessage(ctx context.Context, msg *model.Message, tx *sql.Tx) error {
	const query = `
		update messages
		set state = $1
		where message_id = $2`

	var ec IExecutionContext

	if tx == nil {
		ec = r.db
	} else {
		ec = tx
	}

	_, err := ec.ExecContext(ctx, query, msg.State, msg.MessageId)

	return err
}
