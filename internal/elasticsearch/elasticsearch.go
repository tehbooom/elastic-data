package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/tehbooom/elastic-data/internal/config"
)

type Config struct {
	Client    *elasticsearch.TypedClient
	Ctx       context.Context
	Version   string
	Connected bool
}

func (c *Config) TestConnection() error {
	if c.Connected {
		return nil
	}

	resp, err := c.Client.Core.Info().Do(c.Ctx)
	if err != nil {
		return fmt.Errorf("error connecting to cluster: %w", err)
	}

	log.Debug(resp.ClusterName)
	c.Version = resp.Version.Int
	c.Connected = true

	return nil
}

func SetClient(cfg config.ConfigConnection) (*elasticsearch.TypedClient, error) {
	esConfig := elasticsearch.Config{
		Addresses: cfg.ElasticsearchEndpoints,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	if cfg.Unsafe {
		esConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if cfg.CACert != "" {
		cert, err := os.ReadFile(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate authority %s: %w", cfg.CACert, err)
		}
		esConfig.CACert = cert
	}

	es, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating ES client: %w", err)
	}

	esadd := fmt.Sprintf("ESAddress: %s", esConfig.Addresses)
	esconfig := fmt.Sprintf("ES CONFIG ENDPOINTS: %s", cfg.ElasticsearchEndpoints)

	log.Debug(esadd)
	log.Debug(esconfig)

	return es, nil
}

func (c *Config) BulkRequest(index string, events []map[string]interface{}) error {
	bulk := c.Client.Bulk().Index(index)
	for _, event := range events {
		err := bulk.CreateOp(*&types.CreateOperation{}, event)
		if err != nil {
			return err
		}
	}

	resp, err := bulk.Do(c.Ctx)
	if err != nil {
		return err
	}

	if resp.Errors {
		for _, item := range resp.Items {
			for opType, respItem := range item {
				if respItem.Error != nil {
					log.Printf("Error in %s operation for document %s: %s",
						opType, *respItem.Id_, respItem.Error.Reason)
				}
			}
		}
	}

	return nil
}
