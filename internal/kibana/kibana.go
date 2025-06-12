package kibana

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/config"
	"github.com/tehbooom/go-kibana"
	"github.com/tehbooom/go-kibana/kbapi"
)

type Config struct {
	Client    *kibana.Client
	Ctx       context.Context
	Version   string
	Connected bool
}

func (c *Config) TestConnection() error {
	if c.Connected {
		return nil
	}

	_, err := c.Client.GetRedacted(c.Ctx, &kbapi.GetStatusRequest{})
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error connecting to kibana: %w", err)
	}

	c.Connected = true

	return nil
}

func SetClient(cfg config.ConfigConnection) (*kibana.Client, error) {
	kbConfig := kibana.Config{
		Addresses: cfg.KibanaEndpoints,
	}

	if cfg.APIKey != "" {
		kbConfig.APIKey = cfg.APIKey
	} else {
		kbConfig.Username = cfg.Username
		kbConfig.Password = cfg.Password
	}

	if cfg.Unsafe {
		kbConfig.Transport = &http.Transport{
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
		kbConfig.CACert = cert
	}

	client, err := kibana.NewClient(kbConfig)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create kibana client: %w", err)
	}

	return client, nil
}

func (c *Config) InstallPackage(pkgName string) error {
	_, err := c.Client.EPM.InstallPackageRegistry(c.Ctx, &kbapi.FleetEPMInstallPackageRegistryRequest{
		PackageName: pkgName,
	})

	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to install package %s: %w", pkgName, err)
	}

	return nil
}

func (c *Config) GetInstalledPackages() ([]string, error) {
	resp, err := c.Client.EPM.GetPackagesInstalled(c.Ctx, &kbapi.FleetEPMGetInstalledPackagesRequest{})
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("error getting install packages: %w", err)
	}

	var integrations []string

	for _, integration := range resp.Body.Items {
		integrations = append(integrations, integration.Name)
	}

	return integrations, nil
}
