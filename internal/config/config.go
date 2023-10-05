package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Entity struct {
	Entity         string  `yaml:"entity"`
	Fields         []Field `yaml:"fields"`
	UpdateEndpoint string  `yaml:"update-endpoint"`
	GetEndpoint    string  `yaml:"get-endpoint"`
	ControlFields  string  `yaml:"control-fields,omitempty"`
}

type Config struct {
	RedisAddr         string   `yaml:"redis_addr"`
	DatabaseStructure []Entity `yaml:"database_structure"`
}

func (c *Config) GetEntityConfig(entityName string) *Entity {
	for _, e := range c.DatabaseStructure {
		if e.Entity == entityName {
			return &e
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	//bytes, err := ioutil.ReadFile(filename)
	bytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
