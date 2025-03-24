package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	util "github.com/hankimmy/PtmrBackend/pkg/util"
)

func TestAddPastExperienceToCandidate(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	pastExperience := randomPastExperience("1")

	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience)
	require.NoError(t, err)

	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)

	pastExperiences, found := source["past_experience"].([]interface{})
	require.True(t, found, "Past experience should be found")
	require.Len(t, pastExperiences, 1)
	pastExperienceJSON, err := json.Marshal(pastExperiences[0])
	require.NoError(t, err)

	var body bytes.Buffer
	body.Write(pastExperienceJSON)
	requireBodyMatchPastExperience(t, &body, pastExperience)
	clearIndex(CandidateIdx)
}

func TestAddMultiplePastExperiencesToCandidate(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)
	numExperiences := 3
	var pastExperiences []PastExperience
	for i := 1; i <= numExperiences; i++ {
		pastExperience := randomPastExperience(fmt.Sprintf("%d", i))
		pastExperiences = append(pastExperiences, pastExperience)
		err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience)
		require.NoError(t, err)
	}

	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)

	retrievedPastExperiences, found := source["past_experience"].([]interface{})
	require.True(t, found, "Past experience should be found")
	require.Len(t, retrievedPastExperiences, numExperiences)
	sort.Slice(pastExperiences, func(i, j int) bool {
		return pastExperiences[i].StartDate.Before(pastExperiences[j].StartDate)
	})

	for i, pastExperience := range pastExperiences {
		pastExperienceJSON, err := json.Marshal(retrievedPastExperiences[i])
		require.NoError(t, err)

		var body bytes.Buffer
		body.Write(pastExperienceJSON)
		requireBodyMatchPastExperience(t, &body, pastExperience)
	}
	clearIndex(CandidateIdx)
}

func TestUpdatePastExperienceInCandidate(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	pastExperience := randomPastExperience("1")

	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience)
	require.NoError(t, err)

	pastExperience.JobTitle = "Senior Software Engineer"
	err = esClient.UpdatePastExperienceInCandidate(context.Background(), candidate.UserUid, pastExperience)
	require.NoError(t, err)
	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)

	pastExperiences, found := source["past_experience"].([]interface{})
	require.True(t, found, "Past experience should be found")
	require.Len(t, pastExperiences, 1)

	pastExperienceJSON, err := json.Marshal(pastExperiences[0])
	require.NoError(t, err)

	var body bytes.Buffer
	body.Write(pastExperienceJSON)
	requireBodyMatchPastExperience(t, &body, pastExperience)
	clearIndex(CandidateIdx)
}

func TestDeletePastExperienceFromCandidate(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	pastExperience := randomPastExperience("1")

	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience)
	require.NoError(t, err)

	err = esClient.DeletePastExperienceFromCandidate(context.Background(), candidate.UserUid, pastExperience.ID)
	require.NoError(t, err)

	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)

	pastExperiences, found := source["past_experience"].([]interface{})
	require.True(t, found, "Past experience should be found")
	require.Empty(t, pastExperiences, "Past experience should be deleted")
}

func TestGetPastExperience(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	pastExperience := randomPastExperience("1")

	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience)
	require.NoError(t, err)

	exp, err := esClient.GetPastExperience(context.Background(), candidate.UserUid, pastExperience.ID)
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.Equal(t, pastExperience.ID, exp.ID)
	require.Equal(t, pastExperience.JobTitle, exp.JobTitle)

	exp, err = esClient.GetPastExperience(context.Background(), candidate.UserUid, "non-existent-id")
	require.Error(t, err)
	require.Nil(t, exp)

	clearIndex(CandidateIdx)
}

func TestListPastExperiences(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	pastExperience1 := randomPastExperience("1")
	pastExperience2 := randomPastExperience("2")
	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience1)
	require.NoError(t, err)
	err = esClient.AddPastExperienceToCandidate(context.Background(), candidate.UserUid, pastExperience2)
	require.NoError(t, err)

	experiences, err := esClient.ListPastExperiences(context.Background(), candidate.UserUid)
	require.NoError(t, err)
	require.Len(t, experiences, 2)

	experienceMap := make(map[string]PastExperience)
	for _, exp := range experiences {
		experienceMap[exp.ID] = exp
	}
	require.Contains(t, experienceMap, pastExperience1.ID)
	require.Contains(t, experienceMap, pastExperience2.ID)
	require.Equal(t, pastExperience1.JobTitle, experienceMap[pastExperience1.ID].JobTitle)
	require.Equal(t, pastExperience2.JobTitle, experienceMap[pastExperience2.ID].JobTitle)

	clearIndex(CandidateIdx)
}

func randomPastExperience(id string) PastExperience {
	start, end := RandomDateRange(2024)
	return PastExperience{
		ID:          id,
		Industry:    util.RandomString(10),
		Employer:    util.RandomString(10),
		JobTitle:    util.RandomString(10),
		StartDate:   start,
		EndDate:     end,
		Present:     util.RandomBool(),
		Description: util.RandomString(20),
	}
}

func requireBodyMatchPastExperience(t *testing.T, body *bytes.Buffer, pastExperience PastExperience) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPastExperience PastExperience
	err = json.Unmarshal(data, &gotPastExperience)
	require.NoError(t, err)
	require.Equal(t, pastExperience.ID, gotPastExperience.ID)
	require.Equal(t, pastExperience.Industry, gotPastExperience.Industry)
	require.Equal(t, pastExperience.JobTitle, gotPastExperience.JobTitle)
	require.Equal(t, pastExperience.Employer, gotPastExperience.Employer)
	require.Equal(t, pastExperience.EndDate, gotPastExperience.EndDate)
	require.Equal(t, pastExperience.StartDate, gotPastExperience.StartDate)
	require.Equal(t, pastExperience.Present, gotPastExperience.Present)
	require.Equal(t, pastExperience.Description, gotPastExperience.Description)
}

func RandomDateRange(year int) (time.Time, time.Time) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	randomStartDate := randomDate(startOfYear, endOfYear).UTC().Truncate(24 * time.Hour)
	randomEndDate := randomDate(randomStartDate, endOfYear).UTC().Truncate(24 * time.Hour)

	return randomStartDate, randomEndDate
}

// randomDate generates a random date between the start and end dates
func randomDate(start, end time.Time) time.Time {
	delta := end.Unix() - start.Unix()
	sec := rand.Int63n(delta) + start.Unix()
	return time.Unix(sec, 0)
}
