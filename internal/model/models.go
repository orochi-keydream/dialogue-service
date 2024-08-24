package model

import "time"

type UserId string

type MessageId int64

type ChatId string

type MessageState int32

const (
	MessageStateSent    MessageState = 1
	MessageStatePending MessageState = 2
	MessageStateRemoved MessageState = 3
)

type Message struct {
	MessageId  MessageId
	ChatId     ChatId
	SentAt     time.Time
	FromUserId UserId
	ToUserId   UserId
	Text       string
	State      MessageState
}

type SendMessageCommand struct {
	FromUserId UserId
	ToUserId   UserId
	Text       string
}

type GetMessagesCommand struct {
	FromUserId UserId
	ToUserId   UserId
}

type OutboxMessageType int32

const (
	OutboxMessageTypeAddNewUnreadMessage = 1
)

type OutboxMessage struct {
	Id           int64
	Type         OutboxMessageType
	MessageKey   any
	MessageValue any
	IsSent       bool
}

type AddNewUnreadMessage struct {
	CorrelationId string
	UserId        UserId
	ChatId        ChatId
	MessageId     MessageId
}

type CommitMessageCommand struct {
	CorrelationId string
	MessageId     MessageId
}

type RollbackMessageCommand struct {
	CorrelationId string
	MessageId     MessageId
}
