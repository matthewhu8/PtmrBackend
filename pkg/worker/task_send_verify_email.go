package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	VerificationLink string `json:"verification_link"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal payload")
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	if payload.Name == "" || payload.Email == "" || payload.VerificationLink == "" {
		log.Error().Msg("missing required fields in payload")
		return fmt.Errorf("missing required fields: %w", asynq.SkipRetry) // or handle it differently
	}

	workingDir, err := os.Getwd()
	if err != nil {
		log.Error().Msg("failed to get working directory")
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Construct paths relative to the working directory
	logoPath := filepath.Join(workingDir, "assets", "images", "logo.png")
	iconPath := filepath.Join(workingDir, "assets", "images", "email-icon.png")

	// Validate if file paths exist
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		log.Error().Msg("logo image not found")
		return fmt.Errorf("logo image not found: %s", logoPath)
	}
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		log.Error().Msg("email icon image not found")
		return fmt.Errorf("icon image not found: %s", iconPath)
	}

	// Map to cid references
	attachments := map[string]string{
		"logo":       logoPath,
		"email-icon": iconPath,
	}

	subject := "Welcome to Part Timer"
	content := fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Welcome to Part Timer</title>
		<style>
			@import url('https://fonts.googleapis.com/css2?family=Lexend:wght@300;400;500;700&display=swap');
	
			body {
				font-family: 'Lexend', sans-serif;
				background-color: #f4f4f4;
				margin: 0;
				padding: 0;
			}
			.email-container {
				width: 100%%;
				padding: 20px;
				background-color: #ffffff;
				border-radius: 8px;
				max-width: 600px;
				margin: 0 auto;
				box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);
				text-align: center;
			}
			.logo {
				margin-bottom: 20px;
			}
			.icon {
				margin: 20px 0;
			}
			.email-header {
				font-size: 1.5em;
				font-weight: bold;
				color: #333333;
			}
			.email-subtitle {
				font-size: 1.2em;
				color: #333333;
			}
			.email-content {
				font-size: 1em;
				color: #666666;
				margin-top: 20px;
			}
			.confirm-button {
				background-color: #FD8618;
				color: white !important;
				padding: 10px 20px;
				text-decoration: none;
				border-radius: 25px;
				display: inline-block;
				margin-top: 20px;
				font-size: 1em;
				font-weight: bold;
			}
			.footer {
				font-size: 0.8em;
				color: #aaaaaa;
				margin-top: 20px;
			}
		</style>
	</head>
	<body>
		<div class="email-container">
			<div class="logo">
				<img src="cid:logo" alt="Part Timer Logo" width="150px">
			</div>
			<div class="icon">
				<img src="cid:email-icon" alt="Email Icon" width="150px">
			</div>
			<div class="email-header">
				Hi! %s
			</div>
			<div class="email-subtitle">
				Welcome to Part-Timer
			</div>
			<div class="email-content">
				"Part-timer: Connecting you to the right jobs and talent—faster, smarter, and tailored to your needs!"
			</div>
			<a href="%s" class="confirm-button">Verify your Email</a>
			<div class="footer">
				&copy; 2024 Part-Timer. All rights reserved.
			</div>
		</div>
	</body>
	</html>`, payload.Name, payload.VerificationLink)

	to := []string{payload.Email}

	// Attempt to send the email with attachment
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil, attachments)
	if err != nil {
		// Log detailed error information
		log.Error().Err(err).Msgf("failed to send verify email to: %s", payload.Email)
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().
		Str("task_type", task.Type()).
		Str("email", payload.Email).
		Bytes("payload", task.Payload()).
		Msg("processed task successfully")

	return nil
}
