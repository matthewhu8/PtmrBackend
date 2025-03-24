package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hankimmy/PtmrBackend/pkg/util"

	"github.com/stretchr/testify/require"
)

func createRandomPastExperience(t *testing.T, candidateID int64) PastExperience {
	startDate, endDate := RandomDateRange(2024)
	arg := CreatePastExperienceParams{
		CandidateID: candidateID,
		Industry:    util.RandomString(10),
		Employer:    util.RandomString(10),
		JobTitle:    util.RandomString(10),
		StartDate:   startDate,
		EndDate:     endDate,
		Present:     util.RandomBool(),
		Description: util.RandomString(50),
	}

	pastExperience, err := testStore.CreatePastExperience(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, pastExperience)

	require.Equal(t, arg.CandidateID, pastExperience.CandidateID)
	require.Equal(t, arg.Industry, pastExperience.Industry)
	require.Equal(t, arg.JobTitle, pastExperience.JobTitle)
	require.Equal(t, arg.StartDate, pastExperience.StartDate)
	require.Equal(t, arg.EndDate, pastExperience.EndDate)
	require.Equal(t, arg.Employer, pastExperience.Employer)
	require.Equal(t, arg.Present, pastExperience.Present)
	require.Equal(t, arg.Description, pastExperience.Description)

	require.NotZero(t, pastExperience.ID)
	require.NotZero(t, pastExperience.CreatedAt)

	return pastExperience
}

func TestCreatePastExperience(t *testing.T) {
	candidate := createRandomCandidate(t)
	createRandomPastExperience(t, candidate.ID)
}

func TestGetPastExperience(t *testing.T) {
	candidate := createRandomCandidate(t)
	pastExperience1 := createRandomPastExperience(t, candidate.ID)
	pastExperience2, err := testStore.GetPastExperience(context.Background(), pastExperience1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, pastExperience2)

	require.Equal(t, pastExperience1.ID, pastExperience2.ID)
	require.Equal(t, pastExperience1.CandidateID, pastExperience2.CandidateID)
	require.Equal(t, pastExperience1.Industry, pastExperience2.Industry)
	require.Equal(t, pastExperience1.JobTitle, pastExperience2.JobTitle)
	require.Equal(t, pastExperience1.StartDate, pastExperience2.StartDate)
	require.Equal(t, pastExperience1.Employer, pastExperience2.Employer)
	require.Equal(t, pastExperience1.EndDate, pastExperience2.EndDate)
	require.Equal(t, pastExperience1.Present, pastExperience2.Present)
	require.Equal(t, pastExperience1.Description, pastExperience2.Description)
	require.WithinDuration(t, pastExperience1.CreatedAt, pastExperience2.CreatedAt, time.Second)
}

func TestUpdatePastExperience(t *testing.T) {
	candidate := createRandomCandidate(t)
	pastExperience1 := createRandomPastExperience(t, candidate.ID)

	industry := util.RandomString(10)
	jobTitle := util.RandomString(10)
	description := util.RandomString(50)
	arg := UpdatePastExperienceParams{
		ID: pastExperience1.ID,
		Industry: pgtype.Text{
			String: industry,
			Valid:  true,
		},
		JobTitle: pgtype.Text{
			String: jobTitle,
			Valid:  true,
		},
		Description: pgtype.Text{
			String: description,
			Valid:  true,
		},
	}

	updatedPastExperience, err := testStore.UpdatePastExperience(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedPastExperience)

	require.Equal(t, pastExperience1.ID, updatedPastExperience.ID)
	require.Equal(t, pastExperience1.CandidateID, updatedPastExperience.CandidateID)
	require.Equal(t, industry, updatedPastExperience.Industry)
	require.Equal(t, jobTitle, updatedPastExperience.JobTitle)
	require.Equal(t, pastExperience1.Employer, updatedPastExperience.Employer)
	require.Equal(t, pastExperience1.StartDate, updatedPastExperience.StartDate)
	require.Equal(t, pastExperience1.EndDate, updatedPastExperience.EndDate)
	require.Equal(t, pastExperience1.Present, updatedPastExperience.Present)
	require.Equal(t, description, updatedPastExperience.Description)
	require.WithinDuration(t, pastExperience1.CreatedAt, updatedPastExperience.CreatedAt, time.Second)
}

func TestDeletePastExperience(t *testing.T) {
	candidate := createRandomCandidate(t)
	pastExperience1 := createRandomPastExperience(t, candidate.ID)
	arg := DeletePastExperienceParams{
		ID:          pastExperience1.ID,
		CandidateID: candidate.ID,
	}
	err := testStore.DeletePastExperience(context.Background(), arg)
	require.NoError(t, err)

	pastExperience2, err := testStore.GetPastExperience(context.Background(), pastExperience1.ID)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, pastExperience2)
}

func TestListPastExperiences(t *testing.T) {
	candidate := createRandomCandidate(t)

	for i := 0; i < 10; i++ {
		createRandomPastExperience(t, candidate.ID)
	}

	arg := ListPastExperiencesParams{
		CandidateID: candidate.ID,
		Limit:       5,
		Offset:      0,
	}

	pastExperiences, err := testStore.ListPastExperiences(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, pastExperiences)

	for _, pastExperience := range pastExperiences {
		require.NotEmpty(t, pastExperience)
		require.Equal(t, candidate.ID, pastExperience.CandidateID)
	}
}
