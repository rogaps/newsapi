package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type ESStore interface {
	Search(map[string]interface{}) (*SearchResults, error)
	Create(string, interface{}) error
}

type SearchResults struct {
	Total int    `json:"total"`
	Hits  []*Hit `json:"hits"`
}

type Hit struct {
	ID int `json:"id"`
}

// ESStore represents elasticsearch client on specific index
type esStore struct {
	es        *elasticsearch.Client
	indexName string
}

// NewESSTore connects to elasticsearch and creates a client
func NewESSTore(indexName string, addresses []string, username, password string) (ESStore, error) {
	config := elasticsearch.Config{
		Addresses: addresses,
	}
	if len(username) > 0 {
		config.Username = username
	}
	if len(password) > 0 {
		config.Password = password
	}
	return Connect(indexName, config)
}

// Connect connects to elasticsearch and creates a client
func Connect(indexName string, config elasticsearch.Config) (ESStore, error) {
	es, err := elasticsearch.NewClient(config)
	return &esStore{es, indexName}, err
}

// Create creates a document in elasticsearch
func (s *esStore) Create(docID string, doc interface{}) error {
	payload, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to create document: %s", err)
	}

	ctx := context.Background()
	res, err := esapi.CreateRequest{
		Index:      s.indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(payload),
	}.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("failed to create document: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("failed to create document: %s", err)
		}
		return fmt.Errorf("failed to create document: [%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	return nil
}

func (s *esStore) Search(query map[string]interface{}) (*SearchResults, error) {
	var results SearchResults

	res, err := s.es.Search(
		s.es.Search.WithIndex(s.indexName),
		s.es.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		return &results, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return &results, err
		}
		return &results, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	type envelopeResponse struct {
		Took int
		Hits struct {
			Total struct {
				Value int
			}
			Hits []struct {
				ID         string          `json:"_id"`
				Source     json.RawMessage `json:"_source"`
				Highlights json.RawMessage `json:"highlight"`
				Sort       []interface{}   `json:"sort"`
			}
		}
	}

	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return &results, err
	}

	results.Total = r.Hits.Total.Value

	if len(r.Hits.Hits) < 1 {
		results.Hits = []*Hit{}
		return &results, nil
	}

	for _, hit := range r.Hits.Hits {
		var h Hit
		if err := json.Unmarshal(hit.Source, &h); err != nil {
			return &results, err
		}

		results.Hits = append(results.Hits, &h)
	}

	return &results, nil
}
