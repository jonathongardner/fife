package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type simpleUrl struct {
	host string
	path string
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
	BindHost string           `yaml:"bindHost",json:"bindHost"`
	InfoHost string           `yaml:"infoHost",json:"infoHost"`
	Services []*proxyServices `yaml:"services",json:"services"`
}

func (c config) validate() (err error) {
	names := make(map[string]int)

	for i, s := range c.Services {
		li, ok := names[s.proxyOnHost]
		if ok {
			err = errors.Join(err, fmt.Errorf("duplicate name %d and %d", li, i))
		} else {
			names[s.proxyOnHost] = i
		}
	}

	return
}

type proxyServices struct {
	// Service the server to proxy to
	proxyOnHost string // foo.cool.dev, cool.something.dev, etc
	// the Name of the server to redirect on (the first will be the default)
	proxyToUrl *url.URL // host to redirect on
}

func (ps *proxyServices) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("unsupported type: %v", value.Kind)
	}

	var psI struct {
		// the Name of the server to redirect on (the first will be the default)
		ProxyOn string `yaml:"on"` // foo.cool.dev, cool.something.dev, etc
		// Service the server to proxy to
		ProxyTo string `yaml:"to"` // 192.1.1.1:3000, 192.1.1.1:4000
	}

	err := value.Decode(&psI)
	if err != nil {
		return err
	}

	if psI.ProxyOn == "" {
		return fmt.Errorf("on cant be blank")
	}
	ps.proxyOnHost = psI.ProxyOn

	if psI.ProxyTo == "" {
		return fmt.Errorf("to cant be blank")
	}
	ps.proxyToUrl, err = url.Parse(psI.ProxyTo) // Replace with your backend server
	if err != nil {
		return fmt.Errorf("error parsing to host: %w", err)
	}

	return nil
}
func (ps *proxyServices) MarshalJSON() ([]byte, error) {
	var psI struct {
		// the Name of the server to redirect on (the first will be the default)
		ProxyOn string `json:"on"` // foo.cool.dev, cool.something.dev, etc
		// Service the server to proxy to
		ProxyTo string `json:"to"` // 192.1.1.1:3000, 192.1.1.1:4000
	}
	psI.ProxyOn = ps.proxyOnHost
	psI.ProxyTo = ps.proxyToUrl.String()

	return json.Marshal(&psI)
}

// Proxy returns the http reverse proxy
func (ps *proxyServices) ProxyTo() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(ps.proxyToUrl)
}
