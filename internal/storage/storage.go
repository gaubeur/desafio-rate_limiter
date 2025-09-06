package storage

import (
	"context"
	"time"
)

// Storage é a interface para a estratégia de persistência do rate limiter.
type Storage interface {
	// Increment incrementa o contador para a chave e retorna o novo valor.
	// Se a chave não existir, ela é criada com o valor 1 e a duração fornecida.
	Increment(ctx context.Context, key string, duration time.Duration) (int64, error)
}
