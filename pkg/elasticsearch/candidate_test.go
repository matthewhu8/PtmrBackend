package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func TestIndexCandidate(t *testing.T) {
	candidate := randomCandidate(util.RandomString(5))
	esCandidate := candidateToES(t, candidate)

	err := esClient.IndexCandidate(context.Background(), candidate)
	require.NoError(t, err)

	body := getCandidateDoc(t, candidate.ID)
	requireBodyMatchCandidate(t, &body, esCandidate)
}

func TestIndexCandidateV2(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)

	body := getCandidateDocV2(t, candidate.UserUid)
	requireBodyMatchCandidateV2(t, &body, candidate)
}

func TestUpdateCandidate(t *testing.T) {
	candidate := randomCandidate(util.RandomString(5))
	esCandidate := candidateToES(t, candidate)

	err := esClient.IndexCandidate(context.Background(), candidate)
	require.NoError(t, err)

	// Update candidate
	candidate.FullName = "Jane A. Doe"
	esCandidate.FullName = "Jane A. Doe"
	err = esClient.UpdateCandidate(context.Background(), candidate)
	require.NoError(t, err)

	// Verify candidate is updated
	body := getCandidateDoc(t, candidate.ID)
	requireBodyMatchCandidate(t, &body, esCandidate)
}

func TestUpdateCandidateV2(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)
	candidateMap, err := structToMapV2(candidate)
	require.NoError(t, err)
	a, err := unmarshalTimeAvailabilityJSONV2(candidateMap["time_availability"].(string))
	require.NoError(t, err)
	candidateMap["time_availability"] = a

	// Update candidate
	candidateMap["full_name"] = "Jane A. Doe"
	candidate.FullName = "Jane A. Doe"
	err = esClient.UpdateCandidateV2(context.Background(), candidate.UserUid, candidateMap)
	require.NoError(t, err)

	// Verify candidate is updated
	body := getCandidateDocV2(t, candidate.UserUid)
	requireBodyMatchCandidateV2(t, &body, candidate)
}

func TestPartialUpdateCandidateV2(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)
	updateFields := map[string]interface{}{
		"full_name": "Jane A. Doe",
	}

	// Update candidate
	candidate.FullName = "Jane A. Doe"
	err = esClient.UpdateCandidateV2(context.Background(), candidate.UserUid, updateFields)
	require.NoError(t, err)

	// Verify candidate is updated
	body := getCandidateDocV2(t, candidate.UserUid)
	requireBodyMatchCandidateV2(t, &body, candidate)
}

func TestDeleteCandidateV2(t *testing.T) {
	candidate := randomCandidateV2(util.RandomString(5))
	err := esClient.IndexCandidateV2(context.Background(), candidate)
	require.NoError(t, err)
	err = esClient.DeleteCandidate(context.Background(), candidate.UserUid)
	require.NoError(t, err)
	_, err = esClient.Client.Get().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		Do(context.Background())
	require.Error(t, err)
}

func getCandidateDoc(t *testing.T, candidateID int64) bytes.Buffer {
	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(fmt.Sprintf("%d", candidateID)).
		Do(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, fmt.Sprintf("%d", candidateID), res.Id)
	var body bytes.Buffer
	body.Write(res.Source)
	return body
}

func getCandidateDocV2(t *testing.T, uid string) bytes.Buffer {
	res, err := esClient.Client.Get().
		Index(CandidateIdx).
		Id(uid).
		Do(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, uid, res.Id)
	var body bytes.Buffer
	body.Write(res.Source)
	return body
}

func requireBodyMatchCandidate(t *testing.T, body *bytes.Buffer, candidate esCandidate) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCandidate esCandidate
	err = json.Unmarshal(data, &gotCandidate)
	require.NoError(t, err)

	require.Equal(t, candidate.ID, gotCandidate.ID)
	require.Equal(t, candidate.FullName, gotCandidate.FullName)
	require.Equal(t, candidate.PhoneNumber, gotCandidate.PhoneNumber)
	require.Equal(t, candidate.Education, gotCandidate.Education)
	require.Equal(t, candidate.Location, gotCandidate.Location)
	require.Equal(t, candidate.SkillSet, gotCandidate.SkillSet)
	require.Equal(t, candidate.IndustryOfInterest, gotCandidate.IndustryOfInterest)
	require.Equal(t, candidate.JobPreference, gotCandidate.JobPreference)
	require.Equal(t, candidate.TimeAvailability, gotCandidate.TimeAvailability)
	require.Equal(t, candidate.AccountVerified, gotCandidate.AccountVerified)
	require.Equal(t, candidate.ResumeFile, gotCandidate.ResumeFile)
	require.Equal(t, candidate.ProfilePhoto, gotCandidate.ProfilePhoto)
	require.Equal(t, candidate.Description, gotCandidate.Description)
}

func requireBodyMatchCandidateV2(t *testing.T, body *bytes.Buffer, candidate Candidate) {
	ta, err := unmarshalForTest(candidate.TimeAvailability)
	require.NoError(t, err)

	gotCandidate := struct {
		UserUid            string              `json:"user_uid"`
		FullName           string              `json:"full_name"`
		Email              string              `json:"email"`
		PhoneNumber        string              `json:"phone_number"`
		Education          Education           `json:"education"`
		Location           string              `json:"location"`
		SkillSet           []string            `json:"skill_set"`
		Certificates       []string            `json:"certificates"`
		IndustryOfInterest string              `json:"industry_of_interest"`
		JobPreference      JobPreference       `json:"job_preference"`
		TimeAvailability   []util.Availability `json:"time_availability"`
		ResumeFile         string              `json:"resume_file"`
		ProfilePhoto       string              `json:"profile_photo"`
		Description        string              `json:"description"`
		CreatedAt          time.Time           `json:"created_at"`
	}{
		UserUid:            candidate.UserUid,
		FullName:           candidate.FullName,
		Email:              candidate.Email,
		PhoneNumber:        candidate.PhoneNumber,
		Education:          candidate.Education,
		Location:           candidate.Location,
		SkillSet:           candidate.SkillSet,
		Certificates:       candidate.Certificates,
		IndustryOfInterest: candidate.IndustryOfInterest,
		JobPreference:      candidate.JobPreference,
		TimeAvailability:   ta,
		ResumeFile:         candidate.ResumeFile,
		ProfilePhoto:       candidate.ProfilePhoto,
		Description:        candidate.Description,
		CreatedAt:          candidate.CreatedAt,
	}
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	err = json.Unmarshal(data, &gotCandidate)
	require.NoError(t, err)

	require.Equal(t, candidate.UserUid, gotCandidate.UserUid)
	require.Equal(t, candidate.Email, gotCandidate.Email)
	require.Equal(t, candidate.FullName, gotCandidate.FullName)
	require.Equal(t, candidate.PhoneNumber, gotCandidate.PhoneNumber)
	require.Equal(t, candidate.Education, gotCandidate.Education)
	require.Equal(t, candidate.Location, gotCandidate.Location)
	require.Equal(t, candidate.SkillSet, gotCandidate.SkillSet)
	require.Equal(t, candidate.IndustryOfInterest, gotCandidate.IndustryOfInterest)
	require.Equal(t, candidate.JobPreference, gotCandidate.JobPreference)
	require.Equal(t, ta, gotCandidate.TimeAvailability)
	require.Equal(t, candidate.Certificates, gotCandidate.Certificates)
	require.Equal(t, candidate.ResumeFile, gotCandidate.ResumeFile)
	require.Equal(t, candidate.ProfilePhoto, gotCandidate.ProfilePhoto)
	require.Equal(t, candidate.Description, gotCandidate.Description)
}

func randomCandidate(username string) db.Candidate {
	return db.Candidate{
		ID:                 util.RandomInt(1, 1000),
		Username:           username,
		FullName:           util.RandomString(10),
		PhoneNumber:        util.RandomPhoneNumber(),
		Education:          db.RandomEducation(),
		Location:           util.RandomString(20),
		SkillSet:           []string{util.RandomString(5), util.RandomString(5)},
		IndustryOfInterest: util.RandomString(10),
		JobPreference:      db.RandomJobPref(),
		TimeAvailability:   util.RandomAvailability(),
		AccountVerified:    util.RandomBool(),
		ResumeFile:         util.RandomString(10),
		ProfilePhoto:       util.RandomString(10),
		Description:        util.RandomString(50),
	}
}

func randomCandidateV2(uid string) Candidate {
	return Candidate{
		UserUid:            uid,
		Email:              util.RandomEmail(),
		FullName:           util.RandomString(10),
		PhoneNumber:        util.RandomPhoneNumber(),
		Education:          RandomEducation(),
		Location:           util.RandomString(20),
		Certificates:       []string{util.RandomString(5), util.RandomString(5)},
		SkillSet:           []string{util.RandomString(5), util.RandomString(5)},
		IndustryOfInterest: util.RandomString(10),
		JobPreference:      RandomJobPref(),
		TimeAvailability:   util.RandomAvailability(),
		ResumeFile:         util.RandomString(10),
		ProfilePhoto:       util.RandomString(10),
		Description:        util.RandomString(50),
	}
}

func candidateToES(t *testing.T, candidate db.Candidate) esCandidate {
	ta, err := unmarshalForTest(candidate.TimeAvailability)
	require.NoError(t, err)
	return esCandidate{
		ID:                 candidate.ID,
		Username:           candidate.Username,
		FullName:           candidate.FullName,
		PhoneNumber:        candidate.PhoneNumber,
		Education:          candidate.Education,
		Location:           candidate.Location,
		SkillSet:           candidate.SkillSet,
		IndustryOfInterest: candidate.IndustryOfInterest,
		JobPreference:      candidate.JobPreference,
		TimeAvailability:   ta,
		AccountVerified:    candidate.AccountVerified,
		ResumeFile:         candidate.ResumeFile,
		ProfilePhoto:       candidate.ProfilePhoto,
		Description:        candidate.Description,
		CreatedAt:          candidate.CreatedAt,
	}
}
