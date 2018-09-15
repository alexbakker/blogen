package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/alexbakker/blogen/blog"
	"gopkg.in/yaml.v2"
)

const (
	configFilename = "config.yml"
)

type Config struct {
	Blog blog.Config `yaml:"blog"`
}

func Load(dir string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filepath.Join(dir, configFilename))
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
