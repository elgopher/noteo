package repository

import (
	"os"

	"gopkg.in/yaml.v2"
)

func parse(file string) (*Config, error) {
	c := &Config{}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, c); err != nil {
		return nil, err
	}
	return c, nil
}

type Config struct {
	Editor string `yaml:"editor"`
}

func (r *Config) EditorCommand() string {
	return r.Editor
}
