package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esutil"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const FlushBytes = 4 * 1 << 20

var FlushInterval time.Duration

type bulkActionPayload struct {
	Index string      `json:"_index"`
	Type  string      `json:"_type"`
	ID    json.Number `json:"_id"`
}

type bulkMetadata struct {
	Index  bulkActionPayload `json:"index"`
	Create bulkActionPayload `json:"create"`
	Delete bulkActionPayload `json:"delete"`
	Update bulkActionPayload `json:"update"`
}

func (m bulkMetadata) Action() string {
	if m.Index.Index != "" {
		return "index"
	}

	if m.Create.Index != "" {
		return "create"
	}

	if m.Delete.Index != "" {
		return "delete"
	}

	if m.Update.Index != "" {
		return "update"
	}

	return ""
}

func (m bulkMetadata) Payload() *bulkActionPayload {
	switch m.Action() {
	case "index":
		return &m.Index
	case "create":
		return &m.Create
	case "update":
		return &m.Update
	case "delete":
		return &m.Delete
	}

	return nil
}

type Indexer struct {
	debug     bool
	esClient  *elasticsearch.Client
	indexes   map[string]esutil.BulkIndexer
	indexLock sync.Mutex
}

func NewIndexer(esAddress string) (*Indexer, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:  []string{esAddress},
		MaxRetries: 3,
	})
	if err != nil {
		return nil, err
	}

	indexer := &Indexer{
		esClient: client,
		indexes:  make(map[string]esutil.BulkIndexer),
	}

	return indexer, nil
}

func (i *Indexer) Add(m *bulkMetadata, b []byte) error {
	var index, action string
	if action = m.Action(); action == "" {
		return errors.New("cannot determine required action")
	}

	if index = m.Payload().Index; index == "" {
		return errors.New("empty index")
	}

	indexer, err := i.getESIndexer(index)
	if err != nil {
		return err
	}

	docType := m.Payload().Type
	docId := m.Payload().ID.String()
	docBody := bytes.NewReader(b)

	return indexer.Add(context.Background(), esutil.BulkIndexerItem{
		Index:        index,
		Action:       action,
		DocumentID:   docId,
		DocumentType: docType, // remove on ES upgrade
		Body:         docBody,
	})
}

func (i *Indexer) getESIndexer(index string) (esutil.BulkIndexer, error) {
	i.indexLock.Lock()
	defer i.indexLock.Unlock()

	indexer, ok := i.indexes[index]
	if !ok {
		var err error
		indexer, err = i.newESIndexer(index)
		if err != nil {
			return nil, err
		}

		i.indexes[index] = indexer
		logrus.Infof("created indexer for %s", index)
	}

	return indexer, nil
}

func (i *Indexer) newESIndexer(index string) (esutil.BulkIndexer, error) {
	cfg := esutil.BulkIndexerConfig{
		Index:         index,
		Client:        i.esClient,
		FlushBytes:    FlushBytes,
		FlushInterval: FlushInterval,
		OnError: func(ctx context.Context, err error) {
			logrus.Error(err)
		},
		OnFlushEnd: func(ctx context.Context) {
			logrus.Debugf("indexer for %s was flushed", index)
		},
	}

	if i.debug {
		cfg.DebugLogger = logrus.StandardLogger()
	}

	bi, err := esutil.NewBulkIndexer(cfg)
	if err != nil {
		return nil, err
	}

	return bi, nil
}

func (i *Indexer) Debug() {
	i.debug = true
}
