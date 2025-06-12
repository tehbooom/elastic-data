package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

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
		log.Debug(err)
		return fmt.Errorf("error connecting to elasticsearch cluster: %w", err)
	}

	c.Version = resp.Version.Int
	c.Connected = true

	return nil
}

func SetClient(cfg config.ConfigConnection) (*elasticsearch.TypedClient, error) {
	esConfig := elasticsearch.Config{
		Addresses: cfg.ElasticsearchEndpoints,
	}

	if cfg.APIKey != "" {
		esConfig.APIKey = cfg.APIKey
	} else {
		esConfig.Username = cfg.Username
		esConfig.Password = cfg.Password
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
			log.Debug(err)
			return nil, fmt.Errorf("error reading certificate authority %s: %w", cfg.CACert, err)
		}
		esConfig.CACert = cert
	}

	es, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("error creating ES client: %w", err)
	}

	return es, nil
}

func (c *Config) BulkRequest(index string, events []map[string]interface{}) (time.Duration, error) {
	var duration time.Duration
	bulk := c.Client.Bulk().Index(index)
	for _, event := range events {
		err := bulk.CreateOp(types.CreateOperation{}, event)
		if err != nil {
			log.Debug(err)
			return duration, err
		}
	}

	start := time.Now()
	resp, err := bulk.Do(c.Ctx)
	if err != nil {
		return duration, err
	}
	duration = time.Since(start)

	if resp.Errors {
		for _, item := range resp.Items {
			for opType, respItem := range item {
				if respItem.Error != nil {
					log.Printf("Error in %s operation for document %s: %s",
						opType, *respItem.Id_, *respItem.Error.Reason)
				}
			}
		}
	}

	return duration, nil
}
