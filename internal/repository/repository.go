package repository

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"inventory/internal/config"
)

var ctx = context.Background()

type Repository struct {
	client *redis.Client
	config *config.Config
}

func NewRepository(client *redis.Client, cfg *config.Config) *Repository {
	return &Repository{client: client, config: cfg}
}

func (r *Repository) UpdateEntityInRedis(entityName, guid string, fields map[string]interface{}) error {
	entityConfig := r.config.GetEntityConfig(entityName)
	if entityConfig == nil {
		return errors.New("Invalid entity")
	}

	// Check if the entity already exists
	exists, err := r.client.Exists(ctx, guid).Result()
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	if exists == 0 { // If entity doesn't exist, create a new one
		pipe.HMSet(ctx, guid, fields)
	} else { // If entity exists, update it
		for field, value := range fields {
			pipe.HSet(ctx, guid, field, value)
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *Repository) GetEntityFromRedis(entityName, guid string) (map[string]interface{}, error) {
	entityConfig := r.config.GetEntityConfig(entityName)
	if entityConfig == nil {
		return nil, errors.New("Invalid entity")
	}

	resultString, err := r.client.HGetAll(ctx, guid).Result()
	if err != nil {
		return nil, err
	}

	// Convert map[string]string to map[string]interface{}
	resultInterface := make(map[string]interface{})
	for k, v := range resultString {
		resultInterface[k] = v
	}

	return resultInterface, nil
}
