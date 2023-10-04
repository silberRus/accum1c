package repository

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"inventory/internal/config"
)

var ctx = context.Background()

type Repository struct {
	client            *redis.Client
	databaseStructure []config.Key
}

func NewRepository(client *redis.Client, dbStruct []config.Key) *Repository {
	return &Repository{client: client, databaseStructure: dbStruct}
}

func (r *Repository) UpdateInventoryInRedis(key string, fields map[string]interface{}) error {
	// Проверьте каждое поле на допустимость
	for field := range fields {
		if !r.isValidKeyField(key, field) {
			return errors.New("invalid key or field")
		}
	}
	pipe := r.client.TxPipeline()
	for field, value := range fields {
		pipe.HSet(ctx, key, field, value)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Repository) GetInventoryFromRedis(key, field string) (int, error) {
	// Проверить, существует ли ключ и поле в конфигурации
	if !r.isValidKeyField(key, field) {
		return 0, errors.New("invalid key or field")
	}
	quantity, err := r.client.HGet(ctx, key, field).Int()
	return quantity, err
}

func (r *Repository) isValidKeyField(key, field string) bool {
	for _, k := range r.databaseStructure {
		if k.Key == key {
			for _, f := range k.Fields {
				if f.Name == field {
					return true
				}
			}
		}
	}
	return false
}
