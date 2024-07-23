package api

import (
	"context"

	"github.com/orochi-keydream/dialogue-service/internal/model"
	"github.com/orochi-keydream/dialogue-service/internal/proto/dialogue"
	"github.com/orochi-keydream/dialogue-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DialogueService struct {
	dialogue.UnimplementedDialogueServiceServer

	appService *service.AppService
}

func NewDialogueService(appService *service.AppService) *DialogueService {
	return &DialogueService{
		appService: appService,
	}
}

func (s *DialogueService) GetMessagesV1(ctx context.Context, req *dialogue.GetMessagesV1Request) (*dialogue.GetMessagesV1Response, error) {
	cmd := model.GetMessagesCommand{
		FromUserId: model.UserId(req.FromUserId),
		ToUserId:   model.UserId(req.ToUserId),
	}

	messages, err := s.appService.GetMessages(ctx, cmd)

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	items := make([]*dialogue.GetMessagesV1Response_Message, 0, len(messages))

	for _, message := range messages {
		item := &dialogue.GetMessagesV1Response_Message{
			FromUserId: string(message.FromUserId),
			ToUserId:   string(message.ToUserId),
			Text:       string(message.Text),
		}

		items = append(items, item)
	}

	resp := &dialogue.GetMessagesV1Response{
		Messages: items,
	}

	return resp, nil
}

func (s *DialogueService) SendMessageV1(ctx context.Context, req *dialogue.SendMessageV1Request) (*dialogue.SendMessageV1Response, error) {
	cmd := model.SendMessageCommand{
		FromUserId: model.UserId(req.FromUserId),
		ToUserId:   model.UserId(req.ToUserId),
		Text:       req.Text,
	}

	err := s.appService.SendMessage(ctx, cmd)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dialogue.SendMessageV1Response{}, nil
}
