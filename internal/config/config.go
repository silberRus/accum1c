package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Key struct {
	Key    string  `yaml:"key"`
	Fields []Field `yaml:"fields"`
}

type Endpoint struct {
	UpdateInventory string `yaml:"update_inventory"`
	GetInventory    string `yaml:"get_inventory"`
}

type Config struct {
	RedisAddr         string   `yaml:"redis_addr"`
	Endpoints         Endpoint `yaml:"endpoints"`
	DatabaseStructure []Key    `yaml:"database_structure"`
}

func LoadConfig(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
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
