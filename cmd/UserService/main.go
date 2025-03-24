package main

import (
	"context"
	"os"
	"syscall"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	"github.com/hankimmy/PtmrBackend/pkg/mail"
	"github.com/hankimmy/PtmrBackend/pkg/service"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"UserService/api"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

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
	runTaskProcessor(ctx, waitGroup, dependencies.Config, redisOpt, dependencies.Store, dependencies.ESClient)

	authClient, err := firebase.NewAuthClient(os.Getenv("SERVICE_ACCOUNT_KEY_PATH"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Firebase")
	}

	rateLimiter := api.NewRateLimiter()

	server := api.NewServer(dependencies.Config, dependencies.Store, dependencies.ESClient, taskDistributor, dependencies.TokenMaker, authClient, rateLimiter)
	server.SetupRouter()
	err = server.Start(dependencies.Config.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	esClient elasticsearch.ESClient,
) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, esClient, mailer)

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
