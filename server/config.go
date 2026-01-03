package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
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
	BindHost string           `yaml:"bindHost" json:"bindHost"`
	InfoHost string           `yaml:"infoHost" json:"infoHost"`
	Services []*proxyServices `yaml:"services" json:"services"`
}

func (c config) validate() (err error) {
	names := make(map[string]int)

	for i, s := range c.Services {
		li, ok := names[s.proxyOn.host]
		if ok {
			err = errors.Join(err, fmt.Errorf("duplicate name %d and %d", li, i))
		} else {
			names[s.proxyOn.host] = i
		}
	}

	return
}

// proxyServives is a struct that contains proxy on (the hostname) and proxy to (the request to proxy to).
// It also contains WOL info which can be used to auto send WOL if needed.
type proxyServices struct {
	id string
	// Service the server to proxy to
	proxyOn simpleUrl // foo.cool.dev, cool.something.dev, etc
	// the Name of the server to redirect on (the first will be the default)
	proxyToUrl *url.URL // host to redirect on
	wolInfo    *wol.Info
}

func (ps *proxyServices) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("unsupported type: %v", value.Kind)
	}

	// setup default values
	ps.id = uuid.New().String()

	// Create struct to decode to
	var psI struct {
		// the Name of the server to redirect on (the first will be the default)
		ProxyOn string `yaml:"on"` // foo.cool.dev, cool.something.dev, etc
		// Service the server to proxy to
		ProxyTo string    `yaml:"to"` // http://192.1.1.1:3000, https://192.1.1.1:4000
		WolInfo *wol.Info `yaml:"wol,omitempty"`
	}

	err := value.Decode(&psI)
	if err != nil {
		return err
	}

	// set proxy info
	if psI.ProxyOn == "" {
		return fmt.Errorf("on cant be blank")
	}

	ps.proxyOn.schema = "http://"
	ps.proxyOn.host = strings.TrimPrefix(psI.ProxyOn, "http://")
	if strings.HasPrefix(psI.ProxyOn, "https://") {
		ps.proxyOn.schema = "https://"
		ps.proxyOn.host = strings.TrimPrefix(psI.ProxyOn, "https://")
	}

	if psI.ProxyTo == "" {
		return fmt.Errorf("to cant be blank")
	}
	ps.proxyToUrl, err = url.Parse(psI.ProxyTo) // Replace with your backend server
	if err != nil {
		return fmt.Errorf("error parsing to host: %w", err)
	}

	// set wol info (if needed)
	if psI.WolInfo != nil {
		ps.wolInfo = psI.WolInfo
		if err := psI.WolInfo.Setup(); err != nil {
			return fmt.Errorf("invalid wol info: %w", err)
		}
	}

	return nil
}

// MashalJSON format for json
func (ps *proxyServices) MarshalJSON() ([]byte, error) {
	var psI struct {
		Id string `json:"id"`
		// the Name of the server to redirect on (the first will be the default)
		ProxyOn string `json:"on"` // foo.cool.dev, cool.something.dev, etc
		Name    string `json:"name"`
		// Service the server to proxy to
		ProxyTo string    `json:"to"` // 192.1.1.1:3000, 192.1.1.1:4000
		WolInfo *wol.Info `json:"wol,omitempty"`
	}
	psI.Id = ps.id
	psI.ProxyOn = ps.proxyOn.url()
	psI.Name = ps.proxyOn.host
	psI.ProxyTo = ps.proxyToUrl.String()
	if ps.wolInfo != nil {
		psI.WolInfo = ps.wolInfo
	}

	return json.Marshal(&psI)
}

// Proxy returns the http reverse proxy
func (ps *proxyServices) ProxyTo() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(ps.proxyToUrl)
}
