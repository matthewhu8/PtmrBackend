package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

// createRandomEmployerApplication creates a random employer application for testing.
func createRandomEmployerApplication(t *testing.T, status ApplicationStatus) EmployerApplication {
	employer := createRandomEmployer(t)
	candidate := createRandomCandidate(t)

	arg := CreateEmployerApplicationParams{
		EmployerID:        employer.ID,
		CandidateID:       candidate.ID,
		Message:           util.RandomString(10),
		ApplicationStatus: status,
		CreatedAt:         time.Now(),
	}

	application, err := testStore.CreateEmployerApplication(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, application)

	require.Equal(t, arg.EmployerID, application.EmployerID)
	require.Equal(t, arg.CandidateID, application.CandidateID)
	require.Equal(t, arg.Message, application.Message)
	require.Equal(t, arg.ApplicationStatus, application.ApplicationStatus)
	require.WithinDuration(t, arg.CreatedAt, application.CreatedAt, time.Second)

	return application
}

func TestCreateEmployerApplication(t *testing.T) {
	createRandomEmployerApplication(t, ApplicationStatusPending)
}

func TestGetEmployerApplicationsByCandidate(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusPending)
	applications, err := testStore.GetEmployerApplicationsByCandidate(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
	}
}

func TestUpdateEmployerApplication(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusPending)

	newStatus := ApplicationStatusAccepted
	arg := UpdateEmployerApplicationParams{
		EmployerID:        application.EmployerID,
		CandidateID:       application.CandidateID,
		ApplicationStatus: NullApplicationStatus{ApplicationStatus: newStatus, Valid: true},
	}

	updatedApplication, err := testStore.UpdateEmployerApplication(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, application.EmployerID, updatedApplication.EmployerID)
	require.Equal(t, application.CandidateID, updatedApplication.CandidateID)
	require.Equal(t, newStatus, updatedApplication.ApplicationStatus)
	require.Equal(t, application.Message, updatedApplication.Message)
	require.WithinDuration(t, application.CreatedAt, updatedApplication.CreatedAt, time.Second)
}

func TestUpdateEmployerApplicationStatus(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusSubmitted)
	err := testStore.UpdateEmployerApplicationStatus(context.Background(), UpdateEmployerApplicationStatusParams{
		CandidateID:       application.CandidateID,
		EmployerID:        application.EmployerID,
		ApplicationStatus: ApplicationStatusRejected,
	})
	require.NoError(t, err)
	applications, err := testStore.GetEmployerApplicationsByCandidate(context.Background(), application.CandidateID)
	for _, app := range applications {
		if app.CandidateID == application.CandidateID {
			require.Equal(t, app.ApplicationStatus, ApplicationStatusRejected)
		}
	}
}

func TestDeleteEmployerApplication(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusPending)
	err := testStore.DeleteEmployerApplication(context.Background(), DeleteEmployerApplicationParams{
		EmployerID:  application.EmployerID,
		CandidateID: application.CandidateID,
	})
	require.NoError(t, err)

	applications, err := testStore.GetEmployerApplicationsByCandidate(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.Empty(t, applications)
}
func TestGetEmployerApplicationsByStatusAccepted(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusAccepted)
	applications, err := testStore.GetEmployerApplicationsByStatusAccepted(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusAccepted, app.ApplicationStatus)
	}
}

func TestGetEmployerApplicationsByStatusPending(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusPending)
	applications, err := testStore.GetEmployerApplicationsByStatusPending(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusPending, app.ApplicationStatus)
	}
}

func TestGetEmployerApplicationsByStatusRejected(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusRejected)
	applications, err := testStore.GetEmployerApplicationsByStatusRejected(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusRejected, app.ApplicationStatus)
	}
}

func TestGetEmployerApplicationsByStatusSubmitted(t *testing.T) {
	application := createRandomEmployerApplication(t, ApplicationStatusSubmitted)
	applications, err := testStore.GetEmployerApplicationsByStatusSubmitted(context.Background(), application.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, applications)

	for _, app := range applications {
		require.Equal(t, application.CandidateID, app.CandidateID)
		require.Equal(t, ApplicationStatusSubmitted, app.ApplicationStatus)
	}
}
