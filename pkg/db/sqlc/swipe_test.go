package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func createRandomCandidateSwipe(t *testing.T) CreateCandidateSwipeParams {
	candidate := createRandomCandidate(t)
	arg := CreateCandidateSwipeParams{
		CandidateID: candidate.ID,
		JobID:       util.RandomString(10),
		Swipe:       SwipeAccept,
	}

	err := testStore.CreateCandidateSwipe(context.Background(), arg)
	require.NoError(t, err)

	return arg
}

func createRandomEmployerSwipe(t *testing.T) CreateEmployerSwipesParams {
	employer := createRandomEmployer(t)
	candidate := createRandomCandidate(t)
	arg := CreateEmployerSwipesParams{
		EmployerID:  employer.ID,
		CandidateID: candidate.ID,
		Swipe:       SwipeAccept,
	}

	err := testStore.CreateEmployerSwipes(context.Background(), arg)
	require.NoError(t, err)

	return arg
}

func TestCreateCandidateSwipe(t *testing.T) {
	arg := createRandomCandidateSwipe(t)
	params := GetCandidateSwipeParams{
		CandidateID: arg.CandidateID,
		JobID:       arg.JobID,
	}
	swipe, err := testStore.GetCandidateSwipe(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, swipe)

	require.Equal(t, arg.CandidateID, swipe[0].CandidateID)
	require.Equal(t, arg.JobID, swipe[0].JobID)
	require.Equal(t, arg.Swipe, swipe[0].Swipe)
}

func TestCreateEmployeeSwipe(t *testing.T) {
	arg := createRandomEmployerSwipe(t)
	params := GetEmployerSwipeParams{
		EmployerID:  arg.EmployerID,
		CandidateID: arg.CandidateID,
	}
	swipe, err := testStore.GetEmployerSwipe(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, swipe)

	require.Equal(t, arg.EmployerID, swipe.EmployerID)
	require.Equal(t, arg.CandidateID, swipe.CandidateID)
	require.Equal(t, arg.Swipe, swipe.Swipe)
}

func TestDeleteCandidateSwipe(t *testing.T) {
	arg := createRandomCandidateSwipe(t)

	err := testStore.DeleteCandidateSwipe(context.Background(), DeleteCandidateSwipeParams{
		CandidateID: arg.CandidateID,
		JobID:       arg.JobID,
	})
	require.NoError(t, err)
	params := GetCandidateSwipeParams{
		CandidateID: arg.CandidateID,
		JobID:       arg.JobID,
	}
	swipes, err := testStore.GetCandidateSwipe(context.Background(), params)
	require.NoError(t, err)
	require.Empty(t, swipes, "Expected empty list of swipes after deletion, but got: %v", swipes)
}

func TestDeleteEmployerSwipe(t *testing.T) {
	arg := createRandomEmployerSwipe(t)

	err := testStore.DeleteEmployerSwipe(context.Background(), DeleteEmployerSwipeParams{
		EmployerID:  arg.EmployerID,
		CandidateID: arg.CandidateID,
	})
	require.NoError(t, err)
	params := GetEmployerSwipeParams{
		EmployerID:  arg.EmployerID,
		CandidateID: arg.CandidateID,
	}
	swipes, err := testStore.GetEmployerSwipe(context.Background(), params)
	require.Error(t, err)
	require.Empty(t, swipes, "Expected empty list of swipes after deletion, but got: %v", swipes)
}

func TestGetJobIDsByCandidate(t *testing.T) {
	arg := createRandomCandidateSwipe(t)

	jobIDs, err := testStore.GetJobIDsByCandidate(context.Background(), arg.CandidateID)
	require.NoError(t, err)
	require.NotEmpty(t, jobIDs)

	require.Contains(t, jobIDs, arg.JobID)
}

func TestGetJobIDsByEmployer(t *testing.T) {
	arg := createRandomEmployerSwipe(t)

	candidateIDs, err := testStore.GetCandidateIDsByEmployer(context.Background(), arg.EmployerID)
	require.NoError(t, err)
	require.NotEmpty(t, candidateIDs)

	require.Contains(t, candidateIDs, arg.CandidateID)
}

func createRejectedCandidateSwipe(t *testing.T, candidateID int64) CreateCandidateSwipeParams {
	arg := CreateCandidateSwipeParams{
		CandidateID: candidateID,
		JobID:       util.RandomString(10),
		Swipe:       SwipeReject,
	}

	err := testStore.CreateCandidateSwipe(context.Background(), arg)
	require.NoError(t, err)

	return arg
}

func TestGetRejectedJobIdsByCandidate(t *testing.T) {
	// Test case 1: Candidate with rejected jobs
	t.Run("Candidate with rejected jobs", func(t *testing.T) {
		arg := createRejectedCandidateSwipe(t, createRandomCandidate(t).ID)

		// Call the function to get rejected job IDs
		jobIDs, err := testStore.GetRejectedJobIdsByCandidate(context.Background(), arg.CandidateID)
		require.NoError(t, err)
		require.NotEmpty(t, jobIDs)

		// Check that the job ID is in the result set
		require.Contains(t, jobIDs, arg.JobID)
	})

	// Test case 2: Candidate with no rejected jobs
	t.Run("Candidate with no rejected jobs", func(t *testing.T) {
		arg := createRandomCandidateSwipe(t) // Create a swipe with 'accept'

		// Call the function to get rejected job IDs
		jobIDs, err := testStore.GetRejectedJobIdsByCandidate(context.Background(), arg.CandidateID)
		require.NoError(t, err)
		require.Empty(t, jobIDs)
	})

	// Test case 3: Candidate with multiple rejected jobs
	t.Run("Candidate with multiple rejected jobs", func(t *testing.T) {
		candidate := createRandomCandidate(t)
		arg1 := createRejectedCandidateSwipe(t, candidate.ID)
		arg2 := createRejectedCandidateSwipe(t, candidate.ID)

		// Call the function to get rejected job IDs
		jobIDs, err := testStore.GetRejectedJobIdsByCandidate(context.Background(), candidate.ID)
		require.NoError(t, err)
		require.NotEmpty(t, jobIDs)

		// Check that both job IDs are in the result set
		require.Contains(t, jobIDs, arg1.JobID)
		require.Contains(t, jobIDs, arg2.JobID)
	})
}

func createRejectedEmployerSwipe(t *testing.T, employerID int64) CreateEmployerSwipesParams {
	candidate := createRandomCandidate(t)
	arg := CreateEmployerSwipesParams{
		EmployerID:  employerID,
		CandidateID: candidate.ID,
		Swipe:       SwipeReject,
	}

	err := testStore.CreateEmployerSwipes(context.Background(), arg)
	require.NoError(t, err)

	return arg
}

func TestGetRejectedCandidateIDsByEmployer(t *testing.T) {
	// Test case 1: Candidate with rejected jobs
	t.Run("Candidate with rejected candidates", func(t *testing.T) {
		arg := createRejectedEmployerSwipe(t, createRandomEmployer(t).ID)

		// Call the function to get rejected job IDs
		candidateIDs, err := testStore.GetRejectedCandidateIdsByEmployer(context.Background(), arg.EmployerID)
		require.NoError(t, err)
		require.NotEmpty(t, candidateIDs)

		// Check that the job ID is in the result set
		require.Contains(t, candidateIDs, arg.CandidateID)
	})

	// Test case 2: Candidate with no rejected jobs
	t.Run("Candidate with no rejected jobs", func(t *testing.T) {
		arg := createRandomEmployerSwipe(t) // Create a swipe with 'accept'

		// Call the function to get rejected job IDs
		CandidateIDs, err := testStore.GetRejectedCandidateIdsByEmployer(context.Background(), arg.EmployerID)
		require.NoError(t, err)
		require.Empty(t, CandidateIDs)
	})

	// Test case 3: Candidate with multiple rejected jobs
	t.Run("Candidate with multiple rejected jobs", func(t *testing.T) {
		employer := createRandomEmployer(t)
		arg1 := createRejectedEmployerSwipe(t, employer.ID)
		arg2 := createRejectedEmployerSwipe(t, employer.ID)

		// Call the function to get rejected job IDs
		candidateIDs, err := testStore.GetRejectedCandidateIdsByEmployer(context.Background(), employer.ID)
		require.NoError(t, err)
		require.NotEmpty(t, candidateIDs)

		// Check that both job IDs are in the result set
		require.Contains(t, candidateIDs, arg1.CandidateID)
		require.Contains(t, candidateIDs, arg2.CandidateID)
	})
}
