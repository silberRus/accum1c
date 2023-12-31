package service

import (
	"errors"
	"inventory/internal/config"
	"inventory/internal/repository"
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

	// If there's a control field, check its value
	if entityConfig.ControlFields != "" {
		controlFieldValue, ok := fields[entityConfig.ControlFields]
		if ok {
			if floatValue, ok := controlFieldValue.(float64); ok && floatValue < 0 {
				return errors.New("Control field value cannot be negative")
			}
		}
	}

	// Update the entity in Redis
	guid, ok := fields["guid"].(string)
	if !ok {
		return errors.New("guid is required")
	}
	return s.repo.UpdateEntityInRedis(entityName, guid, fields)
}

func (s *Service) GetEntity(entityName string, guid string) (map[string]interface{}, error) {
	return s.repo.GetEntityFromRedis(entityName, guid)
}
