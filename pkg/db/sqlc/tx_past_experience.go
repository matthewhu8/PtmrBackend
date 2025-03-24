package db

import "context"

type CreatePastExperienceTxParams struct {
	CreatePastExperienceParams
	AfterCreate func(pastExperience PastExperience) error
}

type UpdatePastExperienceTxParams struct {
	UpdatePastExperienceParams
	AfterUpdate func(pastExperience PastExperience) error
}

type DeletePastExperienceTxParams struct {
	DeletePastExperienceParams
	AfterDelete func(candidateID int64, pastExperienceID int64) error
}

type PastExperienceTxResult struct {
	PastExperience PastExperience
}

func (store *SQLStore) CreatePastExperienceTx(ctx context.Context, arg CreatePastExperienceTxParams) (PastExperienceTxResult, error) {
	var result PastExperienceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.PastExperience, err = q.CreatePastExperience(ctx, arg.CreatePastExperienceParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.PastExperience)
	})

	return result, err
}

func (store *SQLStore) UpdatePastExperienceTx(ctx context.Context, arg UpdatePastExperienceTxParams) (PastExperienceTxResult, error) {
	var result PastExperienceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.PastExperience, err = q.UpdatePastExperience(ctx, arg.UpdatePastExperienceParams)
		if err != nil {
			return err
		}

		return arg.AfterUpdate(result.PastExperience)
	})

	return result, err
}

func (store *SQLStore) DeletePastExperienceTx(ctx context.Context, arg DeletePastExperienceTxParams) error {
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		params := arg.DeletePastExperienceParams
		err = q.DeletePastExperience(ctx, params)
		if err != nil {
			return err
		}

		return arg.AfterDelete(params.ID, params.CandidateID)
	})

	return err
}
