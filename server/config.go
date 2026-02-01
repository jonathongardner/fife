package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/jonathongardner/fife/wol"
	yaml "gopkg.in/yaml.v3"
)

type simpleUrl struct {
	schema string
	host   string
}

// return schema + host
func (su simpleUrl) url() string {
	return su.schema + su.host
}

func LoadConfig(file string) (config, error) {
	var config config

	data, err := os.ReadFile(file)
	if err != nil {
		return config, fmt.Errorf("error reading file %w", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling yaml %w", err)
	}

	if config.BindHost == "" {
		return config, fmt.Errorf("bindhost cant be blank")
	}

	if len(config.Services) == 0 {
		return config, fmt.Errorf("service cant be empty")
	}

	if err := config.validate(); err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}

	// TODO: check if services has any duplicate names, if so return error

	return config, nil
}

type config struct {
	BindHost string                    `yaml:"bindHost" json:"bindHost"`
	Services map[string]*proxyServices `yaml:"services" json:"services"`
}

func (c config) validate() (err error) {
	for k, s := range c.Services {
		s.id = k
	}

	return
}

// proxyServives is a struct that contains proxy on (the hostname) and proxy to (the request to proxy to).
// It also contains WOL info which can be used to auto send WOL if needed.
type proxyServices struct {
	id            string   // unqiue name
	redirectToUrl *url.URL // host to redirect on
	wolInfo       wol.Info
}

func (ps *proxyServices) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("unsupported type: %v", value.Kind)
	}

	// Create struct to decode to
	var psI struct {
		RedirectToUrl string   `yaml:"url"` // foo.cool.dev, cool.something.dev, etc
		WolInfo       wol.Info `yaml:"wol,omitempty"`
	}

	err := value.Decode(&psI)
	if err != nil {
		return err
	}

	// set proxy info
	if psI.RedirectToUrl == "" {
		return errors.New("url cant be blank")
	}

	ps.redirectToUrl, err = url.Parse(psI.RedirectToUrl)
	if err != nil {
		return fmt.Errorf("error parsing to host: %w", err)
	}

	ps.wolInfo = psI.WolInfo
	if err := psI.WolInfo.Setup(); err != nil {
		return fmt.Errorf("invalid wol info: %w", err)
	}

	return nil
}

// MashalJSON format for json
func (ps *proxyServices) MarshalJSON() ([]byte, error) {
	var psI struct {
		Id string `json:"id"`
		// the Name of the server to redirect on (the first will be the default)
		RedirectToUrl string   `json:"url"` // foo.cool.dev, cool.something.dev, etc
		WolInfo       wol.Info `json:"wol"`
	}
	psI.Id = ps.id
	psI.RedirectToUrl = ps.redirectToUrl.String()
	psI.WolInfo = ps.wolInfo

	return json.Marshal(&psI)
}
