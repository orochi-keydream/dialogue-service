package app

import (
	"database/sql"
	"fmt"
	"github.com/orochi-keydream/dialogue-service/internal/jobs"
	"github.com/orochi-keydream/dialogue-service/internal/kafka/consumer"
	"github.com/orochi-keydream/dialogue-service/internal/kafka/producer"
	"golang.org/x/net/context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/orochi-keydream/dialogue-service/internal/api"
	"github.com/orochi-keydream/dialogue-service/internal/config"
	"github.com/orochi-keydream/dialogue-service/internal/interceptor"
	"github.com/orochi-keydream/dialogue-service/internal/log"
	"github.com/orochi-keydream/dialogue-service/internal/proto/dialogue"
	"github.com/orochi-keydream/dialogue-service/internal/repository"
	"github.com/orochi-keydream/dialogue-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.LoadConfig()

	addLogger()

	conn := NewConn(cfg.Database)
	dialogueRepository := repository.NewDialogueRepository(conn)
	outboxRepository := repository.NewOutboxRepository(conn)
	commandRepository := repository.NewCommandRepository(conn)
	transactionManager := repository.NewTransactionManager(conn)

	counterCommandProducer, err := producer.NewCounterCommandProducer(cfg.Kafka)

	if err != nil {
		panic(err)
	}

	appService := service.NewAppService(dialogueRepository, outboxRepository, commandRepository, transactionManager)
	outboxService := service.NewOutboxService(outboxRepository, counterCommandProducer, transactionManager)

	wg := &sync.WaitGroup{}

	dialogueCommandConsumer := consumer.NewDialogueCommandConsumer(appService)

	wg.Add(1)
	err = consumer.RunDialogueCommandConsumer(ctx, cfg.Kafka, dialogueCommandConsumer, wg)

	if err != nil {
		panic(err)
	}

	outboxJob := jobs.NewOutboxJob(outboxService)
	outboxJob.Start(ctx)

	grpcDialogueService := api.NewDialogueService(appService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Service.GrpcPort))

	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor,
			interceptor.ErrorInterceptor,
		),
	)

	dialogue.RegisterDialogueServiceServer(server, grpcDialogueService)
	reflection.Register(server)

	go func() {
		err = server.Serve(listener)

		if err != nil {
			panic(err)
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)

	select {
	case <-sigterm:
		server.GracefulStop()
		cancel()
	}

	wg.Wait()

	slog.Info("Gracefully shut down")
}

func addLogger() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	contextHandler := log.NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)
}

func NewConn(cfg config.DatabaseConfig) *sql.DB {
	connStr := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DatabaseName)

	conn, err := sql.Open("pgx", connStr)

	if err != nil {
		panic(err)
	}

	return conn
}
