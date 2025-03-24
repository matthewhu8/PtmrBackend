package elasticsearch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func (c *ESClientImpl) IndexCandidate(ctx context.Context, candidate db.Candidate) error {
	candidateMap, err := structToMap(candidate)
	if err != nil {
		return fmt.Errorf("failed to convert candidate struct to map: %v", err)
	}

	candidateMap["rating"] = util.RandomRating()
	availabilities, err := unmarshalTimeAvailabilityJSON(candidate.TimeAvailability)
	if err != nil {
		return fmt.Errorf("failed to convert candidate time availability from binary to map: %v", err)
	}
	candidateMap["time_availability"] = availabilities

	_, err = c.Client.Index().
		Index(CandidateIdx).
		Id(fmt.Sprintf("%d", candidate.ID)).
		BodyJson(candidateMap).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to index candidate: %v", err)
	}
	return nil
}

func (c *ESClientImpl) IndexCandidateV2(ctx context.Context, candidate Candidate) error {
	candidateMap, err := structToMapV2(candidate)
	if err != nil {
		return fmt.Errorf("failed to convert candidate struct to map: %v", err)
	}
	if ta, ok := candidateMap["time_availability"].(string); ok {
		availabilities, err := unmarshalTimeAvailabilityJSONV2(ta)
		if err != nil {
			return fmt.Errorf("failed to convert candidate time availability from binary to map: %v", err)
		}
		candidateMap["time_availability"] = availabilities
	}
	candidateMap["rating"] = util.RandomRating()

	_, err = c.Client.Index().
		Index(CandidateIdx).
		Id(candidate.UserUid).
		BodyJson(candidateMap).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("%s: %v", ErrIndexFailure, err)
	}
	return nil
}

func (c *ESClientImpl) UpdateCandidate(ctx context.Context, candidate db.Candidate) error {
	candidateMap, err := structToMap(candidate)
	if err != nil {
		return fmt.Errorf("error converting candidate struct to map: %v", err)
	}
	availabilities, err := unmarshalTimeAvailabilityJSON(candidate.TimeAvailability)
	if err != nil {
		return fmt.Errorf("failed to convert candidate time availability from binary to map: %v", err)
	}
	candidateMap["time_availability"] = availabilities

	_, err = c.Client.Update().
		Index(CandidateIdx).
		Id(fmt.Sprintf("%d", candidate.ID)).
		Doc(candidateMap).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error updating candidate: %v", err)
	}

	return nil
}

func (c *ESClientImpl) UpdateCandidateV2(ctx context.Context, userUID string, updateFields map[string]interface{}) error {
	if ta, ok := updateFields["time_availability"].(string); ok {
		availabilities, err := unmarshalTimeAvailabilityJSONV2(ta)
		if err != nil {
			return fmt.Errorf("failed to convert candidate time availability from binary to map: %v", err)
		}
		updateFields["time_availability"] = availabilities
	}

	_, err := c.Client.Update().
		Index(CandidateIdx).
		Id(userUID).
		Doc(updateFields).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("%s: %v", ErrGetFailure, err)
	}

	return nil
}

func (c *ESClientImpl) GetCandidate(ctx context.Context, userUID string) (*Candidate, error) {
	res, err := c.Client.Get().
		Index(CandidateIdx).
		Id(userUID).
		Do(context.Background())
	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %v", ErrGetFailure, err)
	}
	var candidate Candidate
	if err := json.Unmarshal(res.Source, &candidate); err != nil {
		return nil, fmt.Errorf("%s: %v", ErrUnmarshalFailure, err)
	}
	return &candidate, nil
}

func (c *ESClientImpl) DeleteCandidate(ctx context.Context, userUID string) error {
	_, err := c.Client.Delete().
		Index(CandidateIdx).
		Id(userUID).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("%s: %v", ErrDeleteFailure, err)
	}
	return nil
}

func unmarshalTimeAvailabilityJSON(data []byte) ([]map[string]interface{}, error) {
	var timeAvailability []string
	if err := json.Unmarshal(data, &timeAvailability); err != nil {
		return nil, fmt.Errorf("failed to unmarshal time_availability: %v", err)
	}
	var availabilities []map[string]interface{}
	for _, jsonString := range timeAvailability {
		var availability map[string]interface{}
		if err := json.Unmarshal([]byte(jsonString), &availability); err != nil {
			return nil, fmt.Errorf("failed to unmarshal availability JSON string: %v", err)
		}
		availabilities = append(availabilities, availability)
	}
	return availabilities, nil
}

func unmarshalTimeAvailabilityJSONV2(data string) ([]map[string]interface{}, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode time availability data: %v", err)
	}

	var timeAvailability []string
	if err := json.Unmarshal(b, &timeAvailability); err != nil {
		return nil, fmt.Errorf("failed to unmarshal time_availability: %v", err)
	}
	var availabilities []map[string]interface{}
	for _, jsonString := range timeAvailability {
		var availability map[string]interface{}
		if err := json.Unmarshal([]byte(jsonString), &availability); err != nil {
			return nil, fmt.Errorf("failed to unmarshal availability JSON string: %v", err)
		}
		availabilities = append(availabilities, availability)
	}
	return availabilities, nil
}

func unmarshalForTest(data []byte) ([]util.Availability, error) {
	var availabilities []string
	if err := json.Unmarshal(data, &availabilities); err != nil {
		return nil, err
	}

	var result []util.Availability
	for _, jsonString := range availabilities {
		var availability util.Availability
		if err := json.Unmarshal([]byte(jsonString), &availability); err != nil {
			return nil, err
		}
		result = append(result, availability)
	}
	return result, nil
}
