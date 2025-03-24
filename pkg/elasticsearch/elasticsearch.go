package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
)

const (
	CandidateIdx          = "candidates"
	JobIdx                = "jobs"
	CandidateAppIdx       = "candidate_applications"
	EmployerAppIdx        = "employer_applications"
	ErrIndexFailure       = "failed to index"
	ErrGetFailure         = "failed to do ES GET request"
	ErrUpdateFailure      = "failed to update"
	ErrDeleteFailure      = "failed to delete"
	ErrUnmarshalFailure   = "failed to unmarshal"
	ErrNoApplicationFound = "no application found with ID"
)

type ESClient interface {
	IndexJob(job *Job) error
	GetJob(id string) (*Job, error)
	UpdateJob(id string, job *Job) error
	DeleteJob(id string) error
	IndexCandidate(ctx context.Context, candidate db.Candidate) error
	IndexCandidateV2(ctx context.Context, candidate Candidate) error
	UpdateCandidate(ctx context.Context, candidate db.Candidate) error
	UpdateCandidateV2(ctx context.Context, userUID string, updateFields map[string]interface{}) error
	GetCandidate(ctx context.Context, userUID string) (*Candidate, error)
	DeleteCandidate(ctx context.Context, userUID string) error
	AddPastExperienceToCandidate(ctx context.Context, userUID string, pastExperience PastExperience) error
	UpdatePastExperienceInCandidate(ctx context.Context, userUID string, pastExperience PastExperience) error
	DeletePastExperienceFromCandidate(ctx context.Context, userUID string, pastExperienceID string) error
	GetPastExperience(ctx context.Context, userUID, pastExperienceID string) (*PastExperience, error)
	ListPastExperiences(ctx context.Context, userUID string) ([]PastExperience, error)
	IndexCandidateApplication(ctx context.Context, id string, application map[string]interface{}) error
	GetCandidateApplication(ctx context.Context, id string) (map[string]interface{}, error)
	UpdateCandidateApplication(ctx context.Context, id string, application map[string]interface{}) error
	DeleteCandidateApplication(ctx context.Context, id string) error
	IndexEmployerApplication(ctx context.Context, id string, application map[string]interface{}) error
	GetEmployerApplication(ctx context.Context, id string) (map[string]interface{}, error)
	UpdateEmployerApplication(ctx context.Context, id string, application map[string]interface{}) error
	DeleteEmployerApplication(ctx context.Context, id string) error
	SearchJobs(industry, employmentType, title, distance string, candidateLocation GeoPoint) ([]Job, error)
}

type ESClientImpl struct {
	Client *elastic.Client
}

func CreateElasticsearchClient(url string) (ESClient, error) {
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %v", err)
	}
	return &ESClientImpl{Client: client}, nil
}

func structToMap(candidate db.Candidate) (map[string]interface{}, error) {
	data, err := json.Marshal(candidate)
	if err != nil {
		return nil, err
	}

	var candidateMap map[string]interface{}
	err = json.Unmarshal(data, &candidateMap)
	if err != nil {
		return nil, err
	}

	return candidateMap, nil
}

func structToMapV2(candidate Candidate) (map[string]interface{}, error) {
	data, err := json.Marshal(candidate)
	if err != nil {
		return nil, err
	}

	var candidateMap map[string]interface{}
	err = json.Unmarshal(data, &candidateMap)
	if err != nil {
		return nil, err
	}

	return candidateMap, nil
}
