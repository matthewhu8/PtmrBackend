package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	CreateCandidateTx(ctx context.Context, arg CreateCandidateTxParams) (CandidateTxResult, error)
	UpdateCandidateTx(ctx context.Context, arg UpdateCandidateTxParams) (CandidateTxResult, error)
	CreatePastExperienceTx(ctx context.Context, arg CreatePastExperienceTxParams) (PastExperienceTxResult, error)
	UpdatePastExperienceTx(ctx context.Context, arg UpdatePastExperienceTxParams) (PastExperienceTxResult, error)
	DeletePastExperienceTx(ctx context.Context, arg DeletePastExperienceTxParams) error
	CreateCandidateApplicationTx(ctx context.Context, arg CreateCandidateApplicationTxParams) (CandidateAppTxResult, error)
	DeleteApplicationTx(ctx context.Context, arg DeleteApplicationTxParams) error
	UpdateCandidateApplicationStatusTx(ctx context.Context, arg UpdateCandidateApplicationStatusTxParams) error
	UpdateEmployerApplicationStatusTx(ctx context.Context, arg UpdateEmployerApplicationStatusTxParams) error
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
