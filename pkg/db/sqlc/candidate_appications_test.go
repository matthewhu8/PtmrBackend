package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

// createRandomCandidateApplication creates a random candidate application for testing.
func createRandomCandidateApplication(t *testing.T, status ApplicationStatus) CandidateApplication {
	candidate := createRandomCandidate(t)
	employer := createRandomEmployer(t)

	arg := CreateCandidateApplicationParams{
		CandidateID:        candidate.ID,
		EmployerID:         employer.ID,
		ElasticsearchDocID: util.RandomString(10),
		JobDocID:           util.RandomString(5),
		ApplicationStatus:  status,
	}

	application, err := testStore.CreateCandidateApplication(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, application)

	require.Equal(t, arg.CandidateID, application.CandidateID)
	require.Equal(t, arg.EmployerID, application.EmployerID)
	require.Equal(t, arg.ElasticsearchDocID, application.ElasticsearchDocID)
	require.Equal(t, arg.JobDocID, application.JobDocID)
	require.Equal(t, arg.ApplicationStatus, application.ApplicationStatus)

	return application
}

func TestCreateCandidateApplication(t *testing.T) {
	createRandomCandidateApplication(t, ApplicationStatusPending)
}

func TestGetCandidateApplicationsByEmployer(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusPending)
	applications, err := testStore.GetCandidateApplicationsByEmployer(context.Background(), application.EmployerID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.EmployerID, app.EmployerID)
	}
}

func TestUpdateCandidateApplication(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusPending)

	newStatus := ApplicationStatusRejected
	newDocID := util.RandomString(10)
	arg := UpdateCandidateApplicationParams{
		CandidateID:        application.CandidateID,
		EmployerID:         application.EmployerID,
		ApplicationStatus:  NullApplicationStatus{ApplicationStatus: newStatus, Valid: true},
		ElasticsearchDocID: pgtype.Text{String: newDocID, Valid: true},
	}

	updatedApplication, err := testStore.UpdateCandidateApplication(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, application.CandidateID, updatedApplication.CandidateID)
	require.Equal(t, application.EmployerID, updatedApplication.EmployerID)
	require.Equal(t, newStatus, updatedApplication.ApplicationStatus)
	require.Equal(t, newDocID, updatedApplication.ElasticsearchDocID)
	require.WithinDuration(t, application.CreatedAt, updatedApplication.CreatedAt, time.Second)
}

func TestUpdateCandidateApplicationStatus(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusPending)
	err := testStore.UpdateCandidateApplicationStatus(context.Background(), UpdateCandidateApplicationStatusParams{
		CandidateID:       application.CandidateID,
		EmployerID:        application.EmployerID,
		ApplicationStatus: ApplicationStatusRejected,
	})
	require.NoError(t, err)
	applications, err := testStore.GetCandidateApplicationsByEmployer(context.Background(), application.EmployerID)
	for _, app := range applications {
		if app.CandidateID == application.CandidateID {
			require.Equal(t, app.ApplicationStatus, ApplicationStatusRejected)
		}
	}
}

func TestDeleteCandidateApplication(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusPending)
	err := testStore.DeleteCandidateApplication(context.Background(), DeleteCandidateApplicationParams{
		CandidateID: application.CandidateID,
		EmployerID:  application.EmployerID,
	})
	require.NoError(t, err)

	applications, err := testStore.GetCandidateApplicationsByEmployer(context.Background(), application.EmployerID)
	require.NoError(t, err)
	require.Empty(t, applications)
}

func TestGetCandidateApplicationsByStatusAccepted(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusAccepted)
	applications, err := testStore.GetCandidateApplicationsByStatusAccepted(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusAccepted, app.ApplicationStatus)
	}
}

func TestGetCandidateApplicationsByStatusPending(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusPending)
	applications, err := testStore.GetCandidateApplicationsByStatusPending(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusPending, app.ApplicationStatus)
	}
}

func TestGetCandidateApplicationsByStatusRejected(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusRejected)
	applications, err := testStore.GetCandidateApplicationsByStatusRejected(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusRejected, app.ApplicationStatus)
	}
}

func TestGetCandidateApplicationsByStatusSubmitted(t *testing.T) {
	application := createRandomCandidateApplication(t, ApplicationStatusSubmitted)
	applications, err := testStore.GetCandidateApplicationsByStatusSubmitted(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusSubmitted, app.ApplicationStatus)
	}
}
