package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"

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
	cfg := config.LoadConfig()

	addLogger()

	conn := NewConn(cfg)
	repo := repository.NewDialogRepository(conn)

	appService := service.NewAppService(repo)

	grpcDialogueService := api.NewDialogueService(appService)

	listener, err := net.Listen("tcp", ":8084")

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

	err = server.Serve(listener)

	if err != nil {
		panic(err)
	}
}

func addLogger() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	contextHandler := log.NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)
}

func NewConn(cfg config.Config) *sql.DB {
	connStr := fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DatabaseName)

	conn, err := sql.Open("pgx", connStr)

	if err != nil {
		panic(err)
	}

	return conn
}
