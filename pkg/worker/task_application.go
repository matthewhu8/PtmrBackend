package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TaskCreateCandidateApp = "task:create_candidate_app"
	TaskCreateEmployerApp  = "task:create_employer_app"
	TaskDeleteCandidateApp = "task:delete_candidate_app"
	TaskDeleteEmployerApp  = "task:delete_employer_app"
)

type PayloadCreateApplication struct {
	AppDoc map[string]interface{} `json:"app_doc"`
	DocID  string                 `json:"doc_id"`
}

type PayloadDeleteApplication struct {
	DocID string `json:"doc_id"`
}

func (processor *RedisTaskProcessor) processCreateApplicationTask(
	ctx context.Context,
	task *asynq.Task,
	payload PayloadCreateApplication,
	processFunc func(context.Context, string, map[string]interface{}) error,
) error {
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	appDoc := payload.AppDoc
	docID := payload.DocID
	if err := processFunc(ctx, docID, appDoc); err != nil {
		return fmt.Errorf("failed to process task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("create_application_document", docID).Msg("processed task")
	return nil
}

func (processor *RedisTaskProcessor) processDeleteApplicationTask(
	ctx context.Context,
	task *asynq.Task,
	payload PayloadDeleteApplication,
	processFunc func(context.Context, string) error,
) error {
	if err := json.Unmarshal(task.Payload(), payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	docID := payload.DocID
	if err := processFunc(ctx, docID); err != nil {
		return fmt.Errorf("failed to process task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("delete_application_document", docID).Msg("processed task")
	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateCandidateApplication(
	ctx context.Context,
	payload *PayloadCreateApplication,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskCreateCandidateApp, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateEmployerApplication(
	ctx context.Context,
	payload *PayloadCreateApplication,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskCreateEmployerApp, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskDeleteCandidateApplication(
	ctx context.Context,
	payload *PayloadDeleteApplication,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskDeleteCandidateApp, payload, opts...)
}

func (distributor *RedisTaskDistributor) DistributeTaskDeleteEmployerApplication(
	ctx context.Context,
	payload *PayloadDeleteApplication,
	opts ...asynq.Option,
) error {
	return distributor.distributeTask(ctx, TaskDeleteEmployerApp, payload, opts...)
}

func (processor *RedisTaskProcessor) ProcessTaskCreateCandidateApplication(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCreateApplication
	return processor.processCreateApplicationTask(ctx, task, payload, processor.esClient.IndexCandidateApplication)
}

func (processor *RedisTaskProcessor) ProcessTaskCreateEmployerApplication(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCreateApplication
	return processor.processCreateApplicationTask(ctx, task, payload, processor.esClient.IndexEmployerApplication)
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteCandidateApplication(ctx context.Context, task *asynq.Task) error {
	var payload PayloadDeleteApplication
	return processor.processDeleteApplicationTask(ctx, task, payload, processor.esClient.DeleteCandidateApplication)
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteEmployerApplication(ctx context.Context, task *asynq.Task) error {
	var payload PayloadDeleteApplication
	return processor.processDeleteApplicationTask(ctx, task, payload, processor.esClient.DeleteEmployerApplication)
}
