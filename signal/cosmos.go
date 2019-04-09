package signal

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dcos/dcos-signal/config"
	"github.com/segmentio/analytics-go"
)

type CosmosPackages struct {
	AppID              string `json:"appId"`
	PackageInformation struct {
		PackageDefinition struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packageDefinition"`
	} `json:"packageInformation"`
}

// String implements Stringer interface to represent the package in human readable format.
func (c CosmosPackages) String() string {
	return fmt.Sprintf("%s %s", c.PackageInformation.PackageDefinition.Name,
		c.PackageInformation.PackageDefinition.Version)
}

type CosmosReport struct {
	Packages []CosmosPackages `json:"packages"`
}

// String implements Stringer interface to print installed packages in human readable format.
func (c CosmosReport) String() string {
	var pkgs []string
	for _, pkg := range c.Packages {
		pkgs = append(pkgs, pkg.String())
	}
	return strings.Join(pkgs, ", ")
}

// Cosmos implements a Reporter for the cosmos service
type Cosmos struct {
	Report    *CosmosReport
	Endpoints []string
	Method    string
	Headers   map[string]string
	Track     *analytics.Track
	Error     []string
	Name      string
}

func (c *Cosmos) getName() string {
	return c.Name
}

func (c *Cosmos) setReport(body []byte) error {
	if err := json.Unmarshal(body, &c.Report); err != nil {
		return err
	}
	return nil
}

func (c *Cosmos) getReport() interface{} {
	return c.Report
}

func (c *Cosmos) addHeaders(head map[string]string) {
	for k, v := range head {
		c.Headers[k] = v
	}
}

func (c *Cosmos) getHeaders() map[string]string {
	return c.Headers
}

func (c *Cosmos) getEndpoints() []string {
	if len(c.Endpoints) != 1 {
		log.Errorf("Cosmos needs 1 endpoint, got %d", len(c.Endpoints))
	}
	return c.Endpoints
}

func (c *Cosmos) getMethod() string {
	return c.Method
}

func (c *Cosmos) getError() []string {
	return c.Error
}

func (c *Cosmos) appendError(err string) {
	c.Error = append(c.Error, err)
}

func (c *Cosmos) setTrack(config config.Config) error {
	if c.Report == nil {
		return fmt.Errorf("%s report is nil, bailing out.", c.Name)
	}

	log.Infof("Installed cosmos packages: %s", c.Report)
	properties := map[string]interface{}{
		"package_list":       c.Report.Packages,
		"source":             "cluster",
		"customerKey":        config.CustomerKey,
		"environmentVersion": config.DCOSVersion,
		"clusterId":          config.ClusterID,
		"licenseId":          config.LicenseID,
		"variant":            config.DCOSVariant,
		"platform":           config.GenPlatform,
		"provider":           config.GenProvider,
	}

	c.Track = &analytics.Track{
		Event:       "package_list",
		UserId:      config.CustomerKey,
		AnonymousId: config.ClusterID,
		Properties:  properties,
	}
	return nil
}

func (c *Cosmos) getTrack() *analytics.Track {
	return c.Track
}

func (c *Cosmos) sendTrack(config config.Config) error {
	ac := CreateSegmentClient(config.SegmentKey, config.FlagVerbose)
	defer ac.Close()
	err := ac.Track(c.Track)
	return err
}
