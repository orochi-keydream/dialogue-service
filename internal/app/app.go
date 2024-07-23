package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/orochi-keydream/dialogue-service/internal/api"
	"github.com/orochi-keydream/dialogue-service/internal/config"
	"github.com/orochi-keydream/dialogue-service/internal/proto/dialogue"
	"github.com/orochi-keydream/dialogue-service/internal/repository"
	"github.com/orochi-keydream/dialogue-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func Run() {
	cfg := config.LoadConfig()

	conn := NewConn(cfg)
	repo := repository.NewDialogRepository(conn)

	appService := service.NewAppService(repo)

	grpcDialogueService := api.NewDialogueService(appService)

	listener, err := net.Listen("tcp", ":8082")

	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	dialogue.RegisterDialogueServiceServer(server, grpcDialogueService)
	reflection.Register(server)

	err = server.Serve(listener)

	if err != nil {
		log.Fatalln(err)
	}
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
