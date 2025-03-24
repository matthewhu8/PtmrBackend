package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/mail"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskCreateCandidate(ctx context.Context, task *asynq.Task) error
	ProcessTaskUpdateCandidate(ctx context.Context, task *asynq.Task) error
	ProcessTaskAddPastExperience(ctx context.Context, task *asynq.Task) error
	ProcessTaskUpdatePastExperience(ctx context.Context, task *asynq.Task) error
	ProcessTaskDeletePastExperience(ctx context.Context, task *asynq.Task) error
	ProcessTaskCreateCandidateApplication(ctx context.Context, task *asynq.Task) error
	ProcessTaskCreateEmployerApplication(ctx context.Context, task *asynq.Task) error
	ProcessTaskDeleteCandidateApplication(ctx context.Context, task *asynq.Task) error
	ProcessTaskDeleteEmployerApplication(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server   *asynq.Server
	store    db.Store
	esClient elasticsearch.ESClient
	mailer   mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, esClient elasticsearch.ESClient, mailer mail.EmailSender) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server:   server,
		store:    store,
		esClient: esClient,
		mailer:   mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskCreateCandidate, processor.ProcessTaskCreateCandidate)
	mux.HandleFunc(TaskUpdateCandidate, processor.ProcessTaskUpdateCandidate)
	mux.HandleFunc(TaskAddPastExperience, processor.ProcessTaskAddPastExperience)
	mux.HandleFunc(TaskUpdatePastExperience, processor.ProcessTaskUpdatePastExperience)
	mux.HandleFunc(TaskDeletePastExperience, processor.ProcessTaskDeletePastExperience)
	mux.HandleFunc(TaskCreateCandidateApp, processor.ProcessTaskCreateCandidateApplication)
	mux.HandleFunc(TaskCreateEmployerApp, processor.ProcessTaskCreateEmployerApplication)
	mux.HandleFunc(TaskDeleteCandidateApp, processor.ProcessTaskDeleteCandidateApplication)
	mux.HandleFunc(TaskDeleteEmployerApp, processor.ProcessTaskDeleteEmployerApplication)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
