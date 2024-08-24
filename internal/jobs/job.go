package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/orochi-keydream/dialogue-service/internal/service"
)

type OutboxJob struct {
	outboxService *service.OutboxService
}

func NewOutboxJob(outboxService *service.OutboxService) *OutboxJob {
	return &OutboxJob{outboxService}
}

func (oj *OutboxJob) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				oj.Process(ctx)
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

func (oj *OutboxJob) Process(ctx context.Context) {
	err := oj.outboxService.Send(ctx)

	if err != nil {
		slog.Error(err.Error())
	}
}
