package model

import "time"

type UserId string

type MessageId int64

type ChatId string

type Message struct {
	MessageId  MessageId
	ChatId     ChatId
	SentAt     time.Time
	FromUserId UserId
	ToUserId   UserId
	Text       string
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
