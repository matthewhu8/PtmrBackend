package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateCandidate(
		ctx context.Context,
		payload *PayloadCandidate,
		opts ...asynq.Option,
	) error
	DistributeTaskUpdateCandidate(
		ctx context.Context,
		payload *PayloadCandidate,
		opts ...asynq.Option,
	) error
	DistributeTaskAddPastExperience(
		ctx context.Context,
		payload *PayloadPastExperience,
		opts ...asynq.Option,
	) error
	DistributeTaskUpdatePastExperience(
		ctx context.Context,
		payload *PayloadPastExperience,
		opts ...asynq.Option,
	) error
	DistributeTaskDeletePastExperience(
		ctx context.Context,
		payload *PayloadDeletePastExperience,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateCandidateApplication(
		ctx context.Context,
		payload *PayloadCreateApplication,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateEmployerApplication(
		ctx context.Context,
		payload *PayloadCreateApplication,
		opts ...asynq.Option,
	) error
	DistributeTaskDeleteEmployerApplication(
		ctx context.Context,
		payload *PayloadDeleteApplication,
		opts ...asynq.Option,
	) error
	DistributeTaskDeleteCandidateApplication(
		ctx context.Context,
		payload *PayloadDeleteApplication,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

func (distributor *RedisTaskDistributor) distributeTask(
	ctx context.Context,
	taskType string,
	payload interface{},
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(taskType, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}
