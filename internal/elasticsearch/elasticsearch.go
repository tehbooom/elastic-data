package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
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

// detectElasticsearchVersion makes a simple request to detect the ES version
func detectElasticsearchVersion(cfg config.ConfigConnection) (string, error) {
	tempConfig := elasticsearch.Config{
		Addresses: cfg.ElasticsearchEndpoints,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
		},
	}

	if cfg.APIKey != "" {
		tempConfig.APIKey = cfg.APIKey
	} else {
		tempConfig.Username = cfg.Username
		tempConfig.Password = cfg.Password
	}

	if cfg.Unsafe {
		tempConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if cfg.CACert != "" {
		cert, err := os.ReadFile(cfg.CACert)
		if err != nil {
			return "", err
		}
		tempConfig.CACert = cert
	}

	tempClient, err := elasticsearch.NewTypedClient(tempConfig)
	if err != nil {
		return "", err
	}

	resp, err := tempClient.Core.Info().Do(context.Background())
	if err != nil {
		return "", err
	}

	return resp.Version.Int, nil
}

func SetClient(cfg config.ConfigConnection) (*elasticsearch.TypedClient, error) {
	esConfig := elasticsearch.Config{
		Addresses: cfg.ElasticsearchEndpoints,
	}

	version, err := detectElasticsearchVersion(cfg)
	if err != nil {
		log.Warn("Could not detect Elasticsearch version, using default headers", "error", err)
	} else {
		log.Info("Detected Elasticsearch version", "version", version)

		// If version is 8.x, set compatible headers
		if strings.HasPrefix(version, "8.") {
			esConfig.Header = http.Header{
				"Content-Type": []string{"application/json"},
				"Accept":       []string{"application/json"},
			}
		}
	}

	if cfg.APIKey != "" {
		esConfig.APIKey = cfg.APIKey
	} else {
		esConfig.Username = cfg.Username
		esConfig.Password = cfg.Password
	}

	// Configure HTTP transport with optimized settings for high throughput
	transport := &http.Transport{
		MaxIdleConns:        100,              // Increased connection pool
		MaxIdleConnsPerHost: 100,              // Allow more connections per host
		IdleConnTimeout:     90 * time.Second, // Keep connections alive longer
		DisableCompression:  false,            // Enable compression
		DisableKeepAlives:   false,            // Enable keep-alives
	}

	if cfg.Unsafe {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	esConfig.Transport = transport

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
	if len(events) == 0 {
		return 0, nil
	}

	bulk := c.Client.Bulk().Index(index)

	// Batch add operations for better performance
	for _, event := range events {
		err := bulk.CreateOp(types.CreateOperation{}, event)
		if err != nil {
			log.Debug(err)
			return 0, err
		}
	}

	start := time.Now()
	resp, err := bulk.Do(c.Ctx)
	if err != nil {
		return 0, err
	}
	duration := time.Since(start)

	// Only log first few errors to avoid performance impact
	if resp.Errors {
		errorCount := 0
		for _, item := range resp.Items {
			if errorCount >= 5 {
				log.Printf("... and more errors (suppressed for performance)")
				break
			}
			for opType, respItem := range item {
				if respItem.Error != nil {
					log.Printf("Error in %s operation for document %s: %s",
						opType, *respItem.Id_, *respItem.Error.Reason)
					errorCount++
					if errorCount >= 5 {
						break
					}
				}
			}
		}
	}

	return duration, nil
}
