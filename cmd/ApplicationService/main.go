package main

import (
	"context"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/service"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"ApplicationService/api"
)

func main() {
	dependencies, err := service.InitializeService()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}
	defer dependencies.StopFunc()
	redisOpt := asynq.RedisClientOpt{
		Addr: dependencies.Config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	waitGroup, ctx := errgroup.WithContext(dependencies.Ctx)
	runTaskProcessor(ctx, waitGroup, redisOpt, dependencies.Store, dependencies.ESClient)
	server := api.NewServer(dependencies.Config, dependencies.Store, dependencies.ESClient, dependencies.TokenMaker, taskDistributor)
	server.SetupRouter()
	err = server.Start(dependencies.Config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	esClient elasticsearch.ESClient,
) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, esClient, nil)

	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})
}
