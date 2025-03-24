package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
)

const (
	TaskAddPastExperience    = "task:add_past_experience"
	TaskUpdatePastExperience = "task:update_past_experience"
	TaskDeletePastExperience = "task:delete_past_experience"
)

type PayloadPastExperience struct {
	PastExperience es.PastExperience `json:"past_experience"`
	UserUID        string            `json:"user_uid"`
}

type PayloadDeletePastExperience struct {
	PastExperienceID string `json:"past_experience_id"`
	UserUID          string `json:"user_uid"`
}

func (processor *RedisTaskProcessor) processPastExperienceTask(
	ctx context.Context,
	task *asynq.Task,
	payload interface{},
	processFunc func(context.Context, string, es.PastExperience) error,
) error {
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	params := payload.(*PayloadPastExperience)
	if err := processFunc(ctx, params.UserUID, params.PastExperience); err != nil {
		return fmt.Errorf("failed to process candidate: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("candidate", params.UserUID).Msg("processed past experience creation task")
	return nil
}

func (processor *RedisTaskProcessor) processDeletePastExperienceTask(
	ctx context.Context,
	task *asynq.Task,
	payload interface{},
	processFunc func(context.Context, string, string) error,
) error {
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	params := payload.(*PayloadDeletePastExperience)
	if err := processFunc(ctx, params.UserUID, params.PastExperienceID); err != nil {
		return fmt.Errorf("failed to process candidate: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("candidate", params.UserUID).Msg("processed task")
	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskAddPastExperience(
	ctx context.Context,
	payload *PayloadPastExperience,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskAddPastExperience, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskUpdatePastExperience(
	ctx context.Context,
	payload *PayloadPastExperience,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskUpdatePastExperience, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskDeletePastExperience(
	ctx context.Context,
	payload *PayloadDeletePastExperience,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskDeletePastExperience, payload, opts...)
}

func (processor *RedisTaskProcessor) ProcessTaskAddPastExperience(ctx context.Context, task *asynq.Task) error {
	var payload PayloadPastExperience
	return processor.processPastExperienceTask(ctx, task, &payload, processor.esClient.AddPastExperienceToCandidate)
}

func (processor *RedisTaskProcessor) ProcessTaskUpdatePastExperience(ctx context.Context, task *asynq.Task) error {
	var payload PayloadPastExperience
	return processor.processPastExperienceTask(ctx, task, &payload, processor.esClient.UpdatePastExperienceInCandidate)
}

func (processor *RedisTaskProcessor) ProcessTaskDeletePastExperience(ctx context.Context, task *asynq.Task) error {
	var payload PayloadDeletePastExperience
	return processor.processDeletePastExperienceTask(ctx, task, &payload, processor.esClient.DeletePastExperienceFromCandidate)
}
