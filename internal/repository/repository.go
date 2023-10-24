package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"inventory/internal/config"
	"strconv"
	"strings"
)

var ctx = context.Background()

type Repository struct {
	client *redis.Client
	config *config.Config
}

func NewRepository(client *redis.Client, cfg *config.Config) *Repository {
	return &Repository{client: client, config: cfg}
}

func (r *Repository) UpdateEntityInRedis(key string, fields map[string]interface{}) error {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return errors.New("Invalid key format")
	}
	entityName, guid := parts[0], parts[1]

	entityConfig := r.config.GetEntityConfig(entityName)
	if entityConfig == nil {
		return errors.New("Invalid entity")
	}

	pipe := r.client.TxPipeline()

	// Update main hash
	pipe.HMSet(ctx, guid, fields)

	// Update related lists/sets if defined
	if entityConfig.Lists != nil {
		for _, list := range entityConfig.Lists {
			listKey := strings.Replace(list.KeyFormat, "{productGUID}", guid, -1)
			pipe.SAdd(ctx, listKey, guid)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Repository) GetEntityFromRedis(key string) (map[string]interface{}, error) {
	resultString, err := r.client.HGetAll(ctx, key).Result()
	fmt.Printf("Getting entity from Redis with key: %s\n", key)
	fmt.Printf("Redis result string: %v\n", resultString)
	if err != nil {
		return nil, err
	}
	return r.convertRedisResultToEntity(resultString)
}

func (r *Repository) convertRedisResultToEntity(resultString map[string]string) (map[string]interface{}, error) {

	fieldTypes := make(map[string]string)
	for _, entityConfig := range r.config.DatabaseStructure {
		for _, field := range entityConfig.Fields {
			fieldTypes[field.Name] = field.Type
		}
	}

	resultInterface := make(map[string]interface{})
	for k, v := range resultString {
		fieldType, exists := fieldTypes[k]
		if !exists {
			continue
		}

		if r.config.IsEntityName(fieldType) {
			resultInterface[k] = v
			continue
		}

		switch fieldType {
		case "string":
			resultInterface[k] = v
		case "int":
			intValue, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			resultInterface[k] = intValue
		case "float":
			floatValue, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, err
			}
			resultInterface[k] = floatValue
		default:
			resultInterface[k] = v
		}
	}

	fmt.Printf("Converting Redis result to entity. Input string: %v\n", resultString)
	return resultInterface, nil
}

func (r *Repository) GetEntity(entityName, guid string) (map[string]interface{}, error) {
	key := entityName + ":" + guid
	return r.GetEntityFromRedis(key)
}

func (r *Repository) GetEntitiesByFields(entityName string, fields map[string]string) ([]map[string]interface{}, error) {
	keyPrefix := entityName
	for fieldName, fieldValue := range fields {
		keyPrefix += ":" + fieldName + ":" + fieldValue
	}

	// Используем SCAN для получения всех ключей, которые соответствуют нашему шаблону
	var cursor uint64
	var result []map[string]interface{}
	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, keyPrefix+"*", 10).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			entity, err := r.GetEntityFromRedis(key)
			if err != nil {
				return nil, err
			}
			result = append(result, entity)
		}

		if cursor == 0 {
			break
		}
	}

	return result, nil
}
