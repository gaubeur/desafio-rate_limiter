package limiter

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gaubeur/desafio-rate_limiter/internal/storage"
)

func TestRateLimiterMiddleware(t *testing.T) {
	// Configurações de teste
	ipLimit := 5
	ipBlockDuration := 1 * time.Second
	tokenLimits := map[string]Limit{
		"token-A": {MaxRequests: 10, BlockPeriod: 2 * time.Second},
		"token-B": {MaxRequests: 2, BlockPeriod: 3 * time.Second},
	}

	// Um handler simples que sempre retorna 200 OK
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// --- Cenário 1: IP dentro do limite ---
	t.Run("IP within limit", func(t *testing.T) {
		mockStorage := storage.NewMockStorage()
		rl := NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)
		protectedHandler := rl.Middleware(nextHandler)

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		for i := 0; i < ipLimit; i++ {
			protectedHandler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("Esperado status OK, mas obteve %d na requisição %d", rr.Code, i+1)
			}
		}
	})

	// --- Cenário 2: IP excedendo o limite ---
	t.Run("IP exceeds limit", func(t *testing.T) {
		mockStorage := storage.NewMockStorage()
		rl := NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)
		protectedHandler := rl.Middleware(nextHandler)

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rr := httptest.NewRecorder()

		// Faz requisições até o limite
		for i := 0; i < ipLimit; i++ {
			protectedHandler.ServeHTTP(rr, req)
		}

		// A próxima requisição deve ser bloqueada.
		protectedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Esperado status TooManyRequests, mas obteve %d", rr.Code)
		}
	})

	// --- Cenário 3: Token com limite menor que o IP e prioridade ---
	t.Run("Token B overrides IP limit", func(t *testing.T) {
		mockStorage := storage.NewMockStorage()
		rl := NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)
		protectedHandler := rl.Middleware(nextHandler)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "token-B")
		req.RemoteAddr = "192.168.1.3:12345"
		rr := httptest.NewRecorder()

		for i := 0; i < 2; i++ {
			protectedHandler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("Esperado status OK, mas obteve %d na requisição %d com token B", rr.Code, i+1)
			}
		}

		protectedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Esperado status TooManyRequests, mas obteve %d", rr.Code)
		}
	})

	// --- Cenário 4: Requisição com token inválido ou não configurado ---
	t.Run("Invalid token falls back to IP limit", func(t *testing.T) {
		mockStorage := storage.NewMockStorage()
		rl := NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)
		protectedHandler := rl.Middleware(nextHandler)

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "invalid-token")
		req.RemoteAddr = "192.168.1.4:12345"
		rr := httptest.NewRecorder()

		for i := 0; i < ipLimit; i++ {
			protectedHandler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("Esperado status OK, mas obteve %d na requisição %d com token inválido", rr.Code, i+1)
			}
		}

		protectedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Esperado status TooManyRequests, mas obteve %d", rr.Code)
		}
	})

	// --- Cenário 5: Erro no armazenamento ---
	t.Run("Storage error", func(t *testing.T) {
		mockStorage := storage.NewMockStorage()
		rl := NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)
		protectedHandler := rl.Middleware(nextHandler)

		mockStorage.SimulateError = errors.New("simulated storage error")

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		rr := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Esperado status InternalServerError, mas obteve %d", rr.Code)
		}
	})
}
