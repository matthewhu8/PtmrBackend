package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

func (c *ESClientImpl) AddPastExperienceToCandidate(ctx context.Context, userUID string, pastExperience PastExperience) error {
	pastExperienceMap := map[string]interface{}{
		"id":          pastExperience.ID,
		"industry":    pastExperience.Industry,
		"employer":    pastExperience.Employer,
		"job_title":   pastExperience.JobTitle,
		"start_date":  pastExperience.StartDate,
		"end_date":    pastExperience.EndDate,
		"length":      YearsBetween(pastExperience.StartDate, pastExperience.EndDate),
		"present":     pastExperience.Present,
		"description": pastExperience.Description,
	}
	script := `
		if (ctx._source.past_experience == null) {
			ctx._source.past_experience = [params.pastExperience];
		} else {
			ctx._source.past_experience.add(params.pastExperience);
			ctx._source.past_experience.sort((a, b) -> a.start_date.compareTo(b.start_date));
		}
	`
	_, err := c.Client.Update().
		Index(CandidateIdx).
		Id(userUID).
		Script(elastic.NewScript(script).Params(map[string]interface{}{
			"pastExperience": pastExperienceMap,
		})).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error updating candidate: %v", err)
	}
	return nil
}

// UpdatePastExperienceInCandidate updates a specific past experience in a candidate document.
func (c *ESClientImpl) UpdatePastExperienceInCandidate(ctx context.Context, userUID string, pastExperience PastExperience) error {
	var length float64
	if !pastExperience.StartDate.IsZero() && !pastExperience.EndDate.IsZero() {
		length = YearsBetween(pastExperience.StartDate, pastExperience.EndDate)
	} else {
		length = 0.0
	}
	pastExperienceMap := map[string]interface{}{}
	pastExperienceMap["id"] = pastExperience.ID
	if pastExperience.Industry != "" {
		pastExperienceMap["industry"] = pastExperience.Industry
	}
	if pastExperience.JobTitle != "" {
		pastExperienceMap["job_title"] = pastExperience.JobTitle
	}
	if pastExperience.Description != "" {
		pastExperienceMap["description"] = pastExperience.Description
	}
	if !pastExperience.StartDate.IsZero() {
		pastExperienceMap["start_date"] = pastExperience.StartDate
	}
	if !pastExperience.EndDate.IsZero() {
		pastExperienceMap["end_date"] = pastExperience.EndDate
	}
	pastExperienceMap["present"] = pastExperience.Present
	pastExperienceMap["length"] = length

	script := `
		boolean updated = false;
		for (int i = 0; i < ctx._source.past_experience.size(); i++) {
			if (ctx._source.past_experience[i].id == params.pastExperience.id) {
				if (params.pastExperience.containsKey('industry')) {
					ctx._source.past_experience[i].industry = params.pastExperience.industry;
				}
				if (params.pastExperience.containsKey('job_title')) {
					ctx._source.past_experience[i].job_title = params.pastExperience.job_title;
				}
				if (params.pastExperience.containsKey('description')) {
					ctx._source.past_experience[i].description = params.pastExperience.description;
				}
				if (params.pastExperience.containsKey('start_date')) {
					ctx._source.past_experience[i].start_date = params.pastExperience.start_date;
				}
				if (params.pastExperience.containsKey('end_date')) {
					ctx._source.past_experience[i].end_date = params.pastExperience.end_date;
				}
				ctx._source.past_experience[i].length = params.pastExperience.length;
				if (params.pastExperience.containsKey('present')) {
					ctx._source.past_experience[i].present = params.pastExperience.present;
				}
				break;
			}
		}
	`

	params := map[string]interface{}{
		"pastExperience": pastExperienceMap,
	}

	_, err := c.Client.Update().
		Index(CandidateIdx).
		Id(userUID).
		Script(elastic.NewScript(script).Params(params)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error updating past experience: %v", err)
	}

	return nil
}

// DeletePastExperienceFromCandidate deletes a specific past experience from a candidate document.
func (c *ESClientImpl) DeletePastExperienceFromCandidate(ctx context.Context, userUID string, pastExperienceID string) error {
	script := `
		if (ctx._source.past_experience != null) {
			ctx._source.past_experience.removeIf(exp -> exp.id == params.id);
		}
	`
	params := map[string]interface{}{
		"id": pastExperienceID,
	}

	_, err := c.Client.Update().
		Index(CandidateIdx).
		Id(userUID).
		Script(elastic.NewScript(script).Params(params)).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error deleting past experience: %v", err)
	}

	return nil
}

func (c *ESClientImpl) GetPastExperience(ctx context.Context, userUID, pastExperienceID string) (*PastExperience, error) {
	res, err := c.Client.Get().
		Index(CandidateIdx).
		Id(userUID).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate document: %v", err)
	}

	if !res.Found {
		return nil, fmt.Errorf("candidate document not found")
	}

	var source map[string]interface{}
	if err := json.Unmarshal(res.Source, &source); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candidate source: %v", err)
	}

	pastExperiences, ok := source["past_experience"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("past experiences not found in candidate document")
	}

	for _, exp := range pastExperiences {
		expMap := exp.(map[string]interface{})
		if expMap["id"] == pastExperienceID {
			pastExperienceJSON, err := json.Marshal(expMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal past experience: %v", err)
			}

			var pastExperience PastExperience
			if err := json.Unmarshal(pastExperienceJSON, &pastExperience); err != nil {
				return nil, fmt.Errorf("failed to unmarshal past experience: %v", err)
			}

			return &pastExperience, nil
		}
	}

	return nil, fmt.Errorf("past experience with ID %s not found", pastExperienceID)
}

func (c *ESClientImpl) ListPastExperiences(ctx context.Context, userUID string) ([]PastExperience, error) {
	res, err := c.Client.Get().
		Index(CandidateIdx).
		Id(userUID).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate document: %v", err)
	}

	if !res.Found {
		return nil, fmt.Errorf("candidate document not found")
	}

	var source map[string]interface{}
	if err := json.Unmarshal(res.Source, &source); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candidate source: %v", err)
	}

	pastExperiencesRaw, ok := source["past_experience"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("past experiences not found in candidate document")
	}

	pastExperiences := make([]PastExperience, 0, len(pastExperiencesRaw))
	for _, exp := range pastExperiencesRaw {
		expMap := exp.(map[string]interface{})
		pastExperienceJSON, err := json.Marshal(expMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal past experience: %v", err)
		}

		var pastExperience PastExperience
		if err := json.Unmarshal(pastExperienceJSON, &pastExperience); err != nil {
			return nil, fmt.Errorf("failed to unmarshal past experience: %v", err)
		}

		pastExperiences = append(pastExperiences, pastExperience)
	}

	return pastExperiences, nil
}

func YearsBetween(startDate, endDate time.Time) float64 {
	// Calculate the total number of months between the two dates
	totalMonths := int(endDate.Year()-startDate.Year())*12 + int(endDate.Month()-startDate.Month())

	// Calculate the year difference in decimal form
	years := float64(totalMonths) / 12.0

	// Adjust for the remaining days in the month
	startDay := startDate.Day()
	endDay := endDate.Day()
	if endDay < startDay {
		// Calculate the days remaining in the start month
		daysInStartMonth := time.Date(startDate.Year(), startDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		// Add the remaining days to the end day count
		endDay += daysInStartMonth
	}
	dayDifference := float64(endDay-startDay) / 30.0 // Approximate days in a month

	// Add the day difference as a fraction of a year
	years += dayDifference / 12.0

	return years
}
