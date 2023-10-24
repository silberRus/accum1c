package service

import (
	"errors"
	"fmt"
	"inventory/internal/config"
	"inventory/internal/repository"
	"inventory/internal/validation"
)

type Service struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) UpdateEntity(entityName string, fields map[string]interface{}, entityConfig *config.Entity) error {

	// Validate the entity and fields
	if entityConfig == nil {
		return errors.New("Invalid entity")
	}

	fmt.Printf("UpdateEntity inputs: entityName = %s, fields = %v, entityConfig = %v\n", entityName, fields, entityConfig)

	// If there's a control field, check its value
	if entityConfig.ControlFields != "" {
		controlFieldValue, ok := fields[entityConfig.ControlFields]
		if ok {
			if floatValue, ok := controlFieldValue.(float64); ok && floatValue < 0 {
				return errors.New("Control field value cannot be negative")
			}
		}
	}

	// Generate the Redis key
	guid, ok := fields["guid"].(string)
	if !ok {
		return errors.New("guid is required")
	}
	redisKey := entityName + ":" + guid

	// Update the entity in Redis
	return s.repo.UpdateEntityInRedis(redisKey, fields)
}

func (s *Service) GetEntity(entityName string, guid string) (map[string]interface{}, error) {
	if !validation.IsValidGUID(guid) {
		return nil, errors.New("Invalid GUID")
	}
	// Generate the Redis key
	redisKey := entityName + ":" + guid
	fmt.Printf("GetEntity inputs: entityName = %s, redisKey = %s\n", entityName, redisKey)
	return s.repo.GetEntityFromRedis(redisKey)
}
