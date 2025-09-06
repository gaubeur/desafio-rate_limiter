package storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage implementa a interface Storage usando Redis.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage cria uma nova instância de RedisStorage.
func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{client: client}
}

// Increment incrementa o valor de uma chave no Redis.
func (s *RedisStorage) Increment(ctx context.Context, key string, duration time.Duration) (int64, error) {
	// Usa um pipeline para garantir a atomicidade das operações
	pipe := s.client.Pipeline()

	// Incrementa a chave. Se não existir, é criada com valor 1.
	incr := pipe.Incr(ctx, key)

	// Define a expiração da chave. Diferente de EXPIRENX, EXPIRE define
	// a expiração mesmo que a chave já exista.
	// Isso garante que o contador seja resetado após o 'duration' especificado.
	pipe.Expire(ctx, key, duration)

	// Executa o pipeline. As operações de Incr e Expire serão executadas atomicamente.
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	// Retorna o valor incrementado.
	return incr.Val(), nil

}
