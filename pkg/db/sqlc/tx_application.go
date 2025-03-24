package db

import (
	"context"
	"errors"
	"fmt"
)

type CreateCandidateApplicationTxParams struct {
	CreateCandidateApplicationParams
	AppDoc      map[string]interface{}
	AfterCreate func(docID string, appDoc map[string]interface{}) error
}

type DeleteApplicationTxParams struct {
	DeleteCandidateApplicationParams
	DeleteEmployerApplicationParams
	IsEmployer  bool
	DocID       string
	AfterDelete func(docID string) error
}

type UpdateCandidateApplicationStatusTxParams struct {
	UpdateCandidateApplicationStatusParams
	AfterUpdate func(params CreateCandidateSwipeParams) error
}

type UpdateEmployerApplicationStatusTxParams struct {
	UpdateEmployerApplicationStatusParams
	AfterUpdate func(params CreateEmployerSwipesParams) error
}

type CandidateAppTxResult struct {
	CandidateApplication CandidateApplication
}

type EmployerAppTxResult struct {
	EmployerAppResult EmployerApplication
}

func (store *SQLStore) CreateCandidateApplicationTx(ctx context.Context, arg CreateCandidateApplicationTxParams) (CandidateAppTxResult, error) {
	var result CandidateAppTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.CandidateApplication, err = q.CreateCandidateApplication(ctx, arg.CreateCandidateApplicationParams)
		if err != nil {
			return err
		}
		return arg.AfterCreate(result.CandidateApplication.ElasticsearchDocID, arg.AppDoc)
	})
	return result, err
}

func (store *SQLStore) DeleteApplicationTx(ctx context.Context, arg DeleteApplicationTxParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		if arg.IsEmployer {
			err := q.DeleteEmployerApplication(ctx, arg.DeleteEmployerApplicationParams)
			if err != nil {
				return err
			}
			return arg.AfterDelete(arg.DocID)
		} else {
			err := q.DeleteCandidateApplication(ctx, arg.DeleteCandidateApplicationParams)
			if err != nil {
				return err
			}
			return arg.AfterDelete(arg.DocID)
		}
	})
}

func (store *SQLStore) UpdateCandidateApplicationStatusTx(ctx context.Context, arg UpdateCandidateApplicationStatusTxParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		err := q.UpdateCandidateApplicationStatus(ctx, arg.UpdateCandidateApplicationStatusParams)
		if err != nil {
			return err
		}
		appStatusParams := arg.UpdateCandidateApplicationStatusParams
		swipe, err := getSwipe(appStatusParams.ApplicationStatus)
		if err != nil {
			return err
		}
		params := CreateCandidateSwipeParams{
			CandidateID: appStatusParams.CandidateID,
			JobID:       fmt.Sprintf("%d_%d", appStatusParams.CandidateID, appStatusParams.EmployerID),
			Swipe:       swipe,
		}
		return arg.AfterUpdate(params)
	})
}

func (store *SQLStore) UpdateEmployerApplicationStatusTx(ctx context.Context, arg UpdateEmployerApplicationStatusTxParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		err := q.UpdateEmployerApplicationStatus(ctx, arg.UpdateEmployerApplicationStatusParams)
		if err != nil {
			return err
		}
		appStatusParams := arg.UpdateEmployerApplicationStatusParams
		swipe, err := getSwipe(appStatusParams.ApplicationStatus)
		if err != nil {
			return err
		}
		params := CreateEmployerSwipesParams{
			EmployerID:  appStatusParams.EmployerID,
			CandidateID: appStatusParams.CandidateID,
			Swipe:       swipe,
		}
		return arg.AfterUpdate(params)
	})
}

func getSwipe(applicationStatus ApplicationStatus) (Swipe, error) {
	if applicationStatus == ApplicationStatusAccepted {
		return SwipeAccept, nil
	} else if applicationStatus == ApplicationStatusRejected {
		return SwipeReject, nil
	} else {
		return "", errors.New("invalid application status request: use accepted or rejected")
	}
}
