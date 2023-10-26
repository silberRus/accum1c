package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"inventory/internal/config"
	"strings"
)

var ctx = context.Background()

type Repository struct {
	client       *redis.Client
	config       *config.Config
	scriptHashes map[string]string
}

func NewRepository(client *redis.Client, cfg *config.Config) *Repository {
	return &Repository{client: client, config: cfg}
}

func InitializeRedis(r *Repository) {

	conf := config.GlobalConfig
	r.scriptHashes = make(map[string]string)

	for _, accum := range conf.DatabaseStructure.Accums {
		script := GenerateLuaScript(accum)
		sha1, err := r.client.ScriptLoad(ctx, script).Result()
		if err != nil {
			panic(err)
		}
		r.scriptHashes[accum.Accum] = sha1
	}
}

// GenerateLuaScript Генерирует Lua-скрипт для обновления баланса в аккумуляционном регистре
// на основе конфигурационного файла.
func GenerateLuaScript(accumConfig config.Accum) string {

	fieldLines := []string{}
	for _, field := range accumConfig.Fields {
		line := fmt.Sprintf("local %s = redis.call('HGET', KEYS[1], '%s')", field.Name, field.Name)
		fieldLines = append(fieldLines, line)
	}

	balanceLines := []string{}
	for _, balance := range accumConfig.Balance {
		line := fmt.Sprintf("redis.call('HSET', KEYS[1], '%s', %s + ARGV[1])", balance.Name, balance.Name)
		balanceLines = append(balanceLines, line)
	}

	script := strings.Join(fieldLines, "\n") + "\n" + strings.Join(balanceLines, "\n")
	return script
}

func key(entityType string, guid string) string {
	return entityType + ":" + guid
}

// GetEntity получает сущность из Redis.
func GetEntity(r *Repository, entityType string, guid string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key(entityType, guid)).Result()
}

// AddEntity добавляет новую сущность в Redis.
func AddEntity(r *Repository, entityType string, guid string, fields map[string]string) error {
	return r.client.HMSet(ctx, key(entityType, guid), fields).Err()
}

// DeleteEntity удаляет сущность из Redis.
func DeleteEntity(r *Repository, entityType string, guid string) error {
	key := key(entityType, guid)
	if r.client.Exists(ctx, key).Val() == 0 {
		return errors.New("Entity not found guid: " + guid)
	}
	return r.client.Del(ctx, key).Err()
}

// UpdateBalance обновляет баланс в аккумуляционном регистре с помощью Lua-скрипта.
func UpdateBalance(r *Repository, accumType string, fields map[string]string) error {
	sha1, exists := r.scriptHashes[accumType]
	if !exists {
		return errors.New("no script loaded for accum type: " + accumType)
	}

	// Подготовьте аргументы для Lua-скрипта из fields
	args := make([]interface{}, 0, len(fields)*2)
	for field, value := range fields {
		args = append(args, field, value)
	}

	_, err := r.client.EvalSha(ctx, sha1, []string{accumType}, args...).Result()
	return err
}
