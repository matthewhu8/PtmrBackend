package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func requireJSONEqual(t *testing.T, expected, actual []byte) {
	var expectedNormalized, actualNormalized interface{}

	err := json.Unmarshal(expected, &expectedNormalized)
	require.NoError(t, err)

	err = json.Unmarshal(actual, &actualNormalized)
	require.NoError(t, err)

	require.Equal(t, expectedNormalized, actualNormalized)
}

func createRandomCandidate(t *testing.T) Candidate {
	user := createRandomUser(t)

	arg := CreateCandidateParams{
		Username:           user.Username,
		FullName:           util.RandomString(10),
		PhoneNumber:        util.RandomPhoneNumber(),
		Education:          RandomEducation(),
		Location:           util.RandomString(10),
		SkillSet:           []string{util.RandomString(5), util.RandomString(5)},
		Certificates:       []string{util.RandomString(5)},
		IndustryOfInterest: util.RandomString(10),
		JobPreference:      RandomJobPref(),
		TimeAvailability:   util.RandomAvailability(),
		AccountVerified:    util.RandomBool(),
		ResumeFile:         util.RandomString(10),
		ProfilePhoto:       util.RandomString(10),
		Description:        util.RandomString(50),
	}

	candidate, err := testStore.CreateCandidate(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, candidate)

	require.Equal(t, arg.Username, candidate.Username)
	require.Equal(t, arg.FullName, candidate.FullName)
	require.Equal(t, arg.PhoneNumber, candidate.PhoneNumber)
	require.Equal(t, arg.Education, candidate.Education)
	require.Equal(t, arg.Location, candidate.Location)
	require.Equal(t, arg.SkillSet, candidate.SkillSet)
	require.Equal(t, arg.Certificates, candidate.Certificates)
	require.Equal(t, arg.IndustryOfInterest, candidate.IndustryOfInterest)
	require.Equal(t, arg.JobPreference, candidate.JobPreference)
	requireJSONEqual(t, arg.TimeAvailability, candidate.TimeAvailability)
	require.Equal(t, arg.AccountVerified, candidate.AccountVerified)
	require.Equal(t, arg.ResumeFile, candidate.ResumeFile)
	require.Equal(t, arg.ProfilePhoto, candidate.ProfilePhoto)
	require.Equal(t, arg.Description, candidate.Description)

	require.NotZero(t, candidate.ID)
	require.NotZero(t, candidate.CreatedAt)

	return candidate
}

func TestCreateCandidate(t *testing.T) {
	createRandomCandidate(t)
}

func TestGetCandidate(t *testing.T) {
	candidate1 := createRandomCandidate(t)
	candidate2, err := testStore.GetCandidate(context.Background(), candidate1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, candidate2)

	require.Equal(t, candidate1.ID, candidate2.ID)
	require.Equal(t, candidate1.Username, candidate2.Username)
	require.Equal(t, candidate1.FullName, candidate2.FullName)
	require.Equal(t, candidate1.PhoneNumber, candidate2.PhoneNumber)
	require.Equal(t, candidate1.Education, candidate2.Education)
	require.Equal(t, candidate1.Location, candidate2.Location)
	require.Equal(t, candidate1.SkillSet, candidate2.SkillSet)
	require.Equal(t, candidate1.Certificates, candidate2.Certificates)
	require.Equal(t, candidate1.IndustryOfInterest, candidate2.IndustryOfInterest)
	require.Equal(t, candidate1.JobPreference, candidate2.JobPreference)
	requireJSONEqual(t, candidate1.TimeAvailability, candidate2.TimeAvailability)
	require.Equal(t, candidate1.AccountVerified, candidate2.AccountVerified)
	require.Equal(t, candidate1.ResumeFile, candidate2.ResumeFile)
	require.Equal(t, candidate1.ProfilePhoto, candidate2.ProfilePhoto)
	require.Equal(t, candidate1.Description, candidate2.Description)
	require.WithinDuration(t, candidate1.CreatedAt, candidate2.CreatedAt, time.Second)
}

func TestUpdateCandidateFullName(t *testing.T) {
	candidate1 := createRandomCandidate(t)

	newFullName := util.RandomString(10)
	arg := UpdateCandidateParams{
		ID: candidate1.ID,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	}

	updatedCandidate, err := testStore.UpdateCandidate(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, candidate1.ID, updatedCandidate.ID)
	require.Equal(t, newFullName, updatedCandidate.FullName)
	require.Equal(t, candidate1.Username, updatedCandidate.Username)
	require.Equal(t, candidate1.PhoneNumber, updatedCandidate.PhoneNumber)
	require.Equal(t, candidate1.Education, updatedCandidate.Education)
	require.Equal(t, candidate1.Location, updatedCandidate.Location)
	require.Equal(t, candidate1.SkillSet, updatedCandidate.SkillSet)
	require.Equal(t, candidate1.Certificates, updatedCandidate.Certificates)
	require.Equal(t, candidate1.IndustryOfInterest, updatedCandidate.IndustryOfInterest)
	require.Equal(t, candidate1.JobPreference, updatedCandidate.JobPreference)
	requireJSONEqual(t, candidate1.TimeAvailability, updatedCandidate.TimeAvailability)
	require.Equal(t, candidate1.AccountVerified, updatedCandidate.AccountVerified)
	require.Equal(t, candidate1.ResumeFile, updatedCandidate.ResumeFile)
	require.Equal(t, candidate1.ProfilePhoto, updatedCandidate.ProfilePhoto)
	require.Equal(t, candidate1.Description, updatedCandidate.Description)
	require.WithinDuration(t, candidate1.CreatedAt, updatedCandidate.CreatedAt, time.Second)
}

func TestDeleteCandidate(t *testing.T) {
	candidate1 := createRandomCandidate(t)
	err := testStore.DeleteCandidate(context.Background(), candidate1.ID)
	require.NoError(t, err)

	candidate2, err := testStore.GetCandidate(context.Background(), candidate1.ID)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, candidate2)
}

func TestListCandidates(t *testing.T) {
	var lastCandidate Candidate
	for i := 0; i < 10; i++ {
		lastCandidate = createRandomCandidate(t)
	}

	arg := ListCandidatesParams{
		Username: lastCandidate.Username,
		Limit:    5,
		Offset:   0,
	}

	candidates, err := testStore.ListCandidates(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, candidates)

	for _, candidate := range candidates {
		require.NotEmpty(t, candidate)
	}
}
