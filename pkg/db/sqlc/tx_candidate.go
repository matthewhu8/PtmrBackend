package db

import "context"

type CreateCandidateTxParams struct {
	CreateCandidateParams
	AfterCreate func(candidate Candidate) error
}

type UpdateCandidateTxParams struct {
	UpdateCandidateParams
	AfterCreate func(candidate Candidate) error
}

type CandidateTxResult struct {
	Candidate Candidate
}

func (store *SQLStore) CreateCandidateTx(ctx context.Context, arg CreateCandidateTxParams) (CandidateTxResult, error) {
	var result CandidateTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Candidate, err = q.CreateCandidate(ctx, arg.CreateCandidateParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Candidate)
	})

	return result, err
}

func (store *SQLStore) UpdateCandidateTx(ctx context.Context, arg UpdateCandidateTxParams) (CandidateTxResult, error) {
	var result CandidateTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Candidate, err = q.UpdateCandidate(ctx, arg.UpdateCandidateParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Candidate)
	})

	return result, err
}
