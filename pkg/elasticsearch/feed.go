package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
)

const ResultSize = 20

func (c *ESClientImpl) SearchJobs(industry, employmentType, title, distance string,
	candidateLocation GeoPoint) ([]Job, error) {
	query := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery("industry", industry),
			elastic.NewTermQuery("employment_type", employmentType),
			elastic.NewMatchQuery("title", title),
		).
		Filter(
			elastic.NewGeoDistanceQuery("precise_location").
				Lat(candidateLocation.Lat).
				Lon(candidateLocation.Lon).
				Distance(distance),
		)

	res, err := c.Client.Search().
		Index(JobIdx).
		Query(query).
		Size(ResultSize).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to search jobs: %v", err)
	}

	if res.Hits.TotalHits.Value == 0 {
		return nil, nil
	}

	var jobs []Job
	for _, hit := range res.Hits.Hits {
		var job Job
		if err := json.Unmarshal(hit.Source, &job); err != nil {
			return nil, fmt.Errorf("failed to unmarshal job: %v", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}
