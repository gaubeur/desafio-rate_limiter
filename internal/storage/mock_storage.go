package storage

import (
	"context"
	"sync"
	"time"
)

// MockStorage é uma implementação simulada da interface Storage para testes.
type MockStorage struct {
	mu            sync.Mutex
	data          map[string]int64
	exp           map[string]time.Time
	SimulateError error // Campo para simular um erro na chamada Increment.
}

// NewMockStorage cria uma nova instância de MockStorage.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]int64),
		exp:  make(map[string]time.Time),
	}
}

// Increment simula a operação de incremento e expiração do contador.
func (m *MockStorage) Increment(ctx context.Context, key string, duration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Se houver um erro simulado, retorne-o.
	if m.SimulateError != nil {
		return 0, m.SimulateError
	}

	// Verifica se a chave existe e se ainda não expirou.
	if exp, exists := m.exp[key]; exists && exp.After(time.Now()) {
		// A chave existe e é válida, então incrementa o contador.
		m.data[key]++
	} else {
		// A chave não existe ou expirou, então a cria/reinicia.
		m.data[key] = 1
		m.exp[key] = time.Now().Add(duration)
	}

	return m.data[key], nil
}
