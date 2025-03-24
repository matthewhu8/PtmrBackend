package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
)

const (
	TaskCreateCandidate   = "task:create_candidate"
	TaskUpdateCandidate   = "task:update_candidate"
	TaskUpdateCandidateV2 = "task:update_candidate_v2"
)

type PayloadCandidate struct {
	Candidate db.Candidate `json:"candidate"`
}

type PayloadCandidateV2 struct {
	Candidate es.Candidate `json:"candidate"`
}

func (processor *RedisTaskProcessor) processCandidateTask(
	ctx context.Context,
	task *asynq.Task,
	payload interface{},
	processFunc func(context.Context, db.Candidate) error,
) error {
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	candidate := payload.(*PayloadCandidate).Candidate
	if err := processFunc(ctx, candidate); err != nil {
		return fmt.Errorf("failed to process candidate: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("candidate", candidate.Username).Msg("processed task")
	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateCandidate(
	ctx context.Context,
	payload *PayloadCandidate,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskCreateCandidate, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskUpdateCandidate(
	ctx context.Context,
	payload *PayloadCandidate,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskUpdateCandidate, payload, opts...)
}

func (processor *RedisTaskProcessor) ProcessTaskCreateCandidate(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCandidate
	return processor.processCandidateTask(ctx, task, &payload, processor.esClient.IndexCandidate)
}

func (processor *RedisTaskProcessor) ProcessTaskUpdateCandidate(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCandidate
	return processor.processCandidateTask(ctx, task, &payload, processor.esClient.UpdateCandidate)
}
