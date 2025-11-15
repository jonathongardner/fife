package server

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	yaml "gopkg.in/yaml.v3"
)

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
	Services []proxyServices `yaml:"services"`
	BindHost string          `yaml:"bindHost"`
}

func (c config) validate() (err error) {
	names := make(map[string]int)
	defltInt := -1

	for i, s := range c.Services {
		li, ok := names[s.name]
		if ok {
			err = errors.Join(err, fmt.Errorf("duplicate name %d and %d", li, i))
		} else {
			names[s.name] = i
		}

		if s.deflt {
			if defltInt == -1 {
				defltInt = i
			} else {
				err = errors.Join(err, fmt.Errorf("multiple defaults %d and %d", li, i))
			}
		}
	}

	return
}

type proxyServices struct {
	// Host the server to proxy to
	host *url.URL // 192.1.1.1:3000, 192.1.1.1:4000
	// the Name of the server to redirect on (the first will be the default)
	name  string // foo.cool.dev, cool.something.dev, etc
	deflt bool
}

func (ps *proxyServices) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("unsupported type: %v", value.Kind)
	}

	var psI struct {
		// Host the server to proxy to
		Host string `yaml:"host"` // 192.1.1.1:3000, 192.1.1.1:4000
		// the Name of the server to redirect on (the first will be the default)
		Name    string `yaml:"name"` // foo.cool.dev, cool.something.dev, etc
		Default bool   `yaml:"default"`
	}

	err := value.Decode(&psI)
	if err != nil {
		return err
	}

	if psI.Name == "" {
		return fmt.Errorf("name cant be blank")
	}
	ps.name = psI.Name

	if psI.Host == "" {
		return fmt.Errorf("host cant be blank")
	}
	ps.host, err = url.Parse(psI.Host) // Replace with your backend server
	if err != nil {
		return fmt.Errorf("error parsing host: %w", err)
	}

	ps.deflt = psI.Default

	return nil
}
