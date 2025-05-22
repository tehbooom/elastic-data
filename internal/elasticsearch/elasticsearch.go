package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/tehbooom/elastic-data/internal/config"
)

type Config struct {
	Client    *elasticsearch.TypedClient
	Ctx       context.Context
	Version   types.ElasticsearchVersionInfo
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

	c.Version = resp.Version
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

	return es, nil
}
