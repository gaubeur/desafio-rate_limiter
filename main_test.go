package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gaubeur/desafio-rate_limiter/internal/limiter"
	"github.com/gaubeur/desafio-rate_limiter/internal/storage"
)

func TestMainServerAndMiddleware(t *testing.T) {
	// ... (Código para setup do mockStorage e limites)

	// Cria o RateLimiter, injetando o mock.
	mockStorage := storage.NewMockStorage()

	// Define os limites de requisições diretamente no teste
	ipLimit := 5
	ipBlockDuration := 1 * time.Second

	tokenLimits := map[string]limiter.Limit{
		"token-A": {MaxRequests: 10, BlockPeriod: 2 * time.Second},
		"token-B": {MaxRequests: 2, BlockPeriod: 3 * time.Second},
	}

	rl := limiter.NewRateLimiter(mockStorage, ipLimit, ipBlockDuration, tokenLimits)

	// AQUI: Usa a função HomeHandler exportada do pacote main.
	handler := http.HandlerFunc(HomeHandler)
	protectedHandler := rl.Middleware(handler)

	// Cria um servidor de teste.
	ts := httptest.NewServer(protectedHandler)
	defer ts.Close()

	// ... (Cenários de teste)
	t.Run("IP limit test", func(t *testing.T) {
		client := ts.Client()
		for i := 0; i < 5; i++ {
			resp, err := client.Get(ts.URL)
			if err != nil {
				t.Fatalf("Erro na requisição %d: %v", i+1, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Esperado status OK na requisição %d, mas obteve %d", i+1, resp.StatusCode)
			}
			resp.Body.Close()
		}

		resp, err := client.Get(ts.URL)
		if err != nil {
			t.Fatalf("Erro na requisição final: %v", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Esperado status TooManyRequests, mas obteve %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	t.Run("Token B limit test", func(t *testing.T) {
		client := ts.Client()
		req, _ := http.NewRequest("GET", ts.URL, nil)
		req.Header.Set("API_KEY", "token-B")

		for i := 0; i < 2; i++ {
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Erro na requisição %d: %v", i+1, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Esperado status OK na requisição %d, mas obteve %d", i+1, resp.StatusCode)
			}
			resp.Body.Close()
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Erro na requisição final: %v", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Esperado status TooManyRequests, mas obteve %d", resp.StatusCode)
		}
		resp.Body.Close()
	})
}
