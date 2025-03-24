package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"
)

func (c *ESClientImpl) IndexCandidateApplication(ctx context.Context, id string, application map[string]interface{}) error {
	return c.indexDocument(ctx, CandidateAppIdx, id, application)
}

func (c *ESClientImpl) GetCandidateApplication(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.getDocument(ctx, CandidateAppIdx, id)
}

func (c *ESClientImpl) UpdateCandidateApplication(ctx context.Context, id string, application map[string]interface{}) error {
	return c.updateDocument(ctx, CandidateAppIdx, id, application)
}

func (c *ESClientImpl) DeleteCandidateApplication(ctx context.Context, id string) error {
	return c.deleteDocument(ctx, CandidateAppIdx, id)
}

func (c *ESClientImpl) IndexEmployerApplication(ctx context.Context, id string, application map[string]interface{}) error {
	return c.indexDocument(ctx, EmployerAppIdx, id, application)
}

func (c *ESClientImpl) GetEmployerApplication(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.getDocument(ctx, EmployerAppIdx, id)
}

func (c *ESClientImpl) UpdateEmployerApplication(ctx context.Context, id string, application map[string]interface{}) error {
	return c.updateDocument(ctx, EmployerAppIdx, id, application)
}

func (c *ESClientImpl) DeleteEmployerApplication(ctx context.Context, id string) error {
	return c.deleteDocument(ctx, EmployerAppIdx, id)
}

func (c *ESClientImpl) indexDocument(ctx context.Context, index, id string, document map[string]interface{}) error {
	_, err := c.Client.Index().
		Index(index).
		Id(id).
		BodyJson(document).
		Do(ctx)
	if err != nil {
		log.Printf("Indexing document failed: %v", err)
		return fmt.Errorf("%s: %w", ErrIndexFailure, err)
	}
	return nil
}

func (c *ESClientImpl) getDocument(ctx context.Context, index, id string) (map[string]interface{}, error) {
	res, err := c.Client.Get().
		Index(index).
		Id(id).
		Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			return nil, fmt.Errorf("%s", ErrNoApplicationFound)
		}
		log.Printf("Getting document failed: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrGetFailure, err)
	}

	var document map[string]interface{}
	if err := json.Unmarshal(res.Source, &document); err != nil {
		log.Printf("Unmarshalling document failed: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrUnmarshalFailure, err)
	}
	return document, nil
}

func (c *ESClientImpl) updateDocument(ctx context.Context, index, id string, document map[string]interface{}) error {
	_, err := c.Client.Update().
		Index(index).
		Id(id).
		Doc(document).
		Do(ctx)
	if err != nil {
		log.Printf("Updating document failed: %v", err)
		return fmt.Errorf("%s: %w", ErrUpdateFailure, err)
	}
	return nil
}

func (c *ESClientImpl) deleteDocument(ctx context.Context, index, id string) error {
	_, err := c.Client.Delete().
		Index(index).
		Id(id).
		Do(ctx)
	if err != nil {
		log.Printf("Deleting document failed: %v", err)
		return fmt.Errorf("%s: %w", ErrDeleteFailure, err)
	}
	return nil
}
