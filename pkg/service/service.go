package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
)

type Dependencies struct {
	Config     util.Config
	Store      db.Store
	ESClient   elasticsearch.ESClient
	TokenMaker token.Maker
	Ctx        context.Context
	StopFunc   context.CancelFunc
}

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func InitializeService() (*Dependencies, error) {
	configPath := os.Getenv("CONFIG_PATH")
	config, err := util.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	esClient, err := elasticsearch.CreateElasticsearchClient(config.ESSource)
	if err != nil {
		return nil, err
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		stop()
		return nil, err
	}

	store := db.NewStore(connPool)

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	return &Dependencies{
		Config:     config,
		ESClient:   esClient,
		Store:      store,
		TokenMaker: tokenMaker,
		Ctx:        ctx,
		StopFunc:   stop,
	}, nil
}
