package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
)

func (c *ESClientImpl) IndexJob(job *Job) error {
	_, err := c.Client.Index().
		Index(JobIdx).
		Id(job.ID).
		BodyJson(job).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to index job: %v", err)
	}
	return nil
}

func (c *ESClientImpl) GetJob(id string) (*Job, error) {
	res, err := c.Client.Get().
		Index(JobIdx).
		Id(id).
		Do(context.Background())
	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	var job Job
	if err := json.Unmarshal(res.Source, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %v", err)
	}
	return &job, nil
}

func (c *ESClientImpl) UpdateJob(id string, job *Job) error {
	_, err := c.Client.Update().
		Index(JobIdx).
		Id(id).
		Doc(job).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to update job: %v", err)
	}
	return nil
}

func (c *ESClientImpl) DeleteJob(id string) error {
	_, err := c.Client.Delete().
		Index(JobIdx).
		Id(id).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}
	return nil
}
