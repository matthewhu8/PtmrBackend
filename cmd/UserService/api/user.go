package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/hibiken/asynq"
)

type createUserRequest struct {
	Name  string  `json:"name" binding:"required"`
	Email string  `json:"email" binding:"required,email"`
	Role  db.Role `json:"role" binding:"required,oneof=employer candidate"`
}

func (server *Server) sendEmailVerification(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		server.respondWithError(ctx, http.StatusBadRequest, fmt.Errorf("authorization header is not provided"))
		return
	}

	idToken := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := server.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		server.respondWithError(ctx, http.StatusUnauthorized, err)
		return
	}

	uid := token.UID

	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.respondWithError(ctx, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := server.auth.SetUserRole(ctx, uid, string(req.Role)); err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not set user role: %w", err))
		return
	}

	verificationLink, err := server.auth.GenerateEmailVerificationLink(ctx, req.Email, string(req.Role))
	if err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not generate verification link: %w", err))
		return
	}

	taskPayload := &worker.PayloadSendVerifyEmail{
		Name:             req.Name,
		Email:            req.Email,
		VerificationLink: verificationLink,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	if err := server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...); err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not distribute verification task: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":              "User created on firebase, verification email sent, and token needs to be refreshed",
		"refreshTokenRequired": true,
	})
}

// resendEmailRequest struct is the same as createUserRequest but only requires the email
type resendEmailRequest struct {
	Email string  `json:"email" binding:"required,email"`
	Role  db.Role `json:"role" binding:"required,oneof=employer candidate"`
}

// resendEmailVerification handles the resend verification email logic
func (server *Server) resendEmailVerification(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		server.respondWithError(ctx, http.StatusBadRequest, fmt.Errorf("authorization header is not provided"))
		return
	}

	idToken := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := server.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		server.respondWithError(ctx, http.StatusUnauthorized, err)
		return
	}

	uid := token.UID

	// Parse request for the email
	var req resendEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.respondWithError(ctx, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	// Optionally: Check if email is already verified to avoid resending
	user, err := server.auth.GetUserByEmail(ctx, req.Email)
	if err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not retrieve user: %w", err))
		return
	}

	if user.EmailVerified {
		server.respondWithError(ctx, http.StatusBadRequest, fmt.Errorf("email already verified"))
		return
	}

	// Rate limiting check (optional)
	lastSent, found := server.rateLimiter.GetLastSent(uid) // Example: use a rate limiter (in-memory store or Redis)
	if found && time.Since(lastSent) < time.Minute {
		server.respondWithError(ctx, http.StatusTooManyRequests, fmt.Errorf("please wait before resending the email"))
		return
	}

	// Generate new email verification link
	verificationLink, err := server.auth.GenerateEmailVerificationLink(ctx, req.Email, string(req.Role))
	if err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not generate verification link: %w", err))
		return
	}

	// Update rate limiter
	server.rateLimiter.SetLastSent(uid, time.Now()) // Store the time the email was last sent

	// Create payload for the task
	taskPayload := &worker.PayloadSendVerifyEmail{
		Name:             user.DisplayName, // Retrieve the user name from Firebase if needed
		Email:            req.Email,
		VerificationLink: verificationLink,
	}

	// Define options (can adjust delay and retries as needed)
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	// Distribute the email verification task
	if err := server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...); err != nil {
		server.respondWithError(ctx, http.StatusInternalServerError, fmt.Errorf("could not distribute verification task: %w", err))
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Verification email resent successfully.",
	})
}

// Error Handling Functions

func (server *Server) respondWithError(ctx *gin.Context, statusCode int, err error) {
	ctx.JSON(statusCode, errorResponse(err))
}
