package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/kv"
	"github.com/cloudflare/cloudflare-go/v2/option"
)

// type APIInterface interface {
// 	DeleteWorkersKVEntries(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkersKVEntriesParams) (cloudflare.Response, error)
// 	DeleteWorkersKVEntry(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkersKVEntryParams) (cloudflare.Response, error)
// 	GetWorkersKV(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetWorkersKVParams) ([]byte, error)
// 	ListWorkersKVKeys(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVsParams) (cloudflare.ListStorageKeysResponse, error)
// 	WriteWorkersKVEntry(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WriteWorkersKVEntryParams) (cloudflare.Response, error)
// }

type Storage struct {
	// api         APIInterface
	api         *cloudflare.Client
	email       string
	accountID   string
	namespaceID string
}

func New(config ...Config) *Storage {

	cfg := configDefault(config...)

	api := cloudflare.NewClient(
		option.WithAPIToken(cfg.Key),
	)

	// if cfg.Key == "test" {
	// 	api := &TestModule{
	// 		baseUrl: "http://localhost:8787",
	// 	}

	// 	storage := &Storage{
	// 		api:         api,
	// 		email:       "example@cloudflare.org",
	// 		accountID:   "dummy-ID",
	// 		namespaceID: "dummy-ID",
	// 	}

	// 	return storage
	// }

	storage := &Storage{
		api:         api,
		email:       cfg.Email,
		accountID:   cfg.AccountID,
		namespaceID: cfg.NamespaceID,
	}

	return storage
}

func (s *Storage) Get(key string) (*string, error) {

	resp, err := s.api.KV.Namespaces.Values.Get(context.TODO(), s.namespaceID, key, kv.NamespaceValueGetParams{
		AccountID: cloudflare.F(s.accountID),
	})

	if err != nil {
		log.Println("Error occur in GetWorkersKV")
		return nil, err
	}

	return resp, nil
}

func (s *Storage) Set(key string, val []byte, exp time.Duration) error {

	_, err := s.api.KV.Namespaces.Values.Get(context.Background(), s.accountID, key, kv.NamespaceValueGetParams{
		AccountID: cloudflare.F(s.accountID),
	})

	if err != nil {
		log.Println("Error occur in WriteWorkersKVEntry")
		return err
	}

	return nil
}

func (s *Storage) Delete(key string) error {

	_, err := s.api.KV.Namespaces.Values.Delete(context.TODO(), s.accountID, key, kv.NamespaceValueDeleteParams{
		AccountID: cloudflare.F(s.accountID),
		Body:      map[string]interface{}{},
	})

	if err != nil {
		log.Println("Error occur in WriteWorkersKVEntry")
		return err
	}

	return nil
}

func (s *Storage) Reset() error {

	var (
		cursor string
		keys   []string
	)

	for {
		resp, err := s.api.KV.Namespaces.Keys.List(context.TODO(), s.namespaceID, kv.NamespaceKeyListParams{
			AccountID: cloudflare.F(s.accountID),
			Cursor:    cloudflare.F(cursor),
			Limit:     cloudflare.F(10.000000),
			Prefix:    cloudflare.F("My-Prefix"),
		})

		if err != nil {
			log.Println("Error occur in ListWorkersKVKeys")
			return err
		}

		keys = make([]string, len(resp.Result))

		for _, element := range resp.Result {
			name := element.Name
			keys = append(keys, name)
		}

		_, err = s.api.KV.Namespaces.Bulk.Delete(context.TODO(), s.accountID, kv.NamespaceBulkDeleteParams{
			AccountID: cloudflare.F(s.accountID),
			Body:      keys,
		})

		if err != nil {
			log.Println("Error occur in DeleteWorker")
			return err
		}

		if len(resp.ResultInfo.Cursor) == 0 {
			log.Println("No keys left in the namespace")
			break
		}

		cursor = resp.ResultInfo.Cursor
	}

	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Conn() cloudflare.Client {
	return *s.api
}
