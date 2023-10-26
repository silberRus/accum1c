package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Accum struct {
	Accum          string  `yaml:"entity"`
	Fields         []Field `yaml:"fields"`
	Balance        []Field `yaml:"balance"`
	UpdateEndpoint string  `yaml:"update-endpoint"`
	GetEndpoint    string  `yaml:"get-endpoint"`
}

type Entity struct {
	Entity         string  `yaml:"entity"`
	Fields         []Field `yaml:"fields"`
	UpdateEndpoint string  `yaml:"update-endpoint"`
	GetEndpoint    string  `yaml:"get-endpoint"`
	ControlFields  string  `yaml:"control-fields,omitempty"`
	Lists          []struct {
		Name      string `yaml:"name"`
		KeyFormat string `yaml:"key-format"`
	}
}

type Config struct {
	RedisAddr         string `yaml:"redis_addr"`
	DatabaseStructure struct {
		Entitys []Entity `yaml:"entities"`
		Accums  []Accum  `yaml:"accums"`
	}
}

var GlobalConfig = initializeConfig()

func (c *Config) GetEntityConfig(entityName string) *Entity {
	for _, e := range c.DatabaseStructure.Entitys {
		if e.Entity == entityName {
			return &e
		}
	}
	return nil
}

func (c *Config) IsEntityName(name string) bool {
	for _, entity := range c.DatabaseStructure.Entitys {
		if entity.Entity == name {
			return true
		}
	}
	return false
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func initializeConfig() *Config {

	name := "config.yaml"
	paths := []string{".", "..", "../..", "../../.."}

	for _, path := range paths {
		if fileExists(path + "/" + name) {
			return LoadConfig(path + "/" + name)
		}
	}
	panic(fmt.Sprintf("Config file %s not found", name))
}

func LoadConfig(filename string) *Config {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config file: %v", err))
	}

	var c Config
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal config: %v", err))
	}

	return &c
}
