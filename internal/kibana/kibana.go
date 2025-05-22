package kibana

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

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

	_, err := c.Client.Status.GetRedacted(c.Ctx, &kbapi.GetStatusRequest{})
	if err != nil {
		return fmt.Errorf("error connecting to cluster: %w", err)
	}

	c.Connected = true

	return nil
}

func SetClient(cfg config.ConfigConnection) (*kibana.Client, error) {
	kbConfig := kibana.Config{
		Addresses: cfg.KibanaEndpoints,
		Username:  cfg.Username,
		Password:  cfg.Password,
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
			return nil, fmt.Errorf("error reading certificate authority %s: %w", cfg.CACert, err)
		}
		kbConfig.CACert = cert
	}

	client, err := kibana.NewClient(kbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kibana client: %w", err)
	}

	return client, nil
}

func (c *Config) InstallPackage(pkgName, pkgVersion string) error {
	_, err := c.Client.Fleet.EPM.InstallPackageRegistry(c.Ctx, &kbapi.FleetEPMInstallPackageRegistryRequest{
		PackageName:    pkgName,
		PackageVersion: &pkgVersion,
	})
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", pkgName, err)
	}

	return nil
}
