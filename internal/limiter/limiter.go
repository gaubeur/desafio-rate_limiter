package limiter

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gaubeur/desafio-rate_limiter/internal/storage"
)

// Limit define as regras de limitação.
type Limit struct {
	MaxRequests int
	BlockPeriod time.Duration
}

// RateLimiter é o middleware principal.
type RateLimiter struct {
	storage     storage.Storage
	defaultIP   Limit
	tokenLimits map[string]Limit
}

// NewRateLimiter cria uma nova instância de RateLimiter.
func NewRateLimiter(s storage.Storage, ipLimit int, ipBlock time.Duration, tokenLimits map[string]Limit) *RateLimiter {
	return &RateLimiter{
		storage: s,
		defaultIP: Limit{
			MaxRequests: ipLimit,
			BlockPeriod: ipBlock,
		},
		tokenLimits: tokenLimits,
	}
}

// Middleware é a função que atua como o middleware de rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tenta obter o token do cabeçalho
		token := r.Header.Get("API_KEY")

		// Define a chave e o limite a serem usados
		var key string
		var limit Limit

		// Se um token foi fornecido e tem um limite configurado
		if token != "" {
			if tokenLimit, ok := rl.tokenLimits[token]; ok {
				key = fmt.Sprintf("token:%s", token)
				limit = tokenLimit
			}
		}

		// Se não há token ou o token não tem um limite configurado, usa o IP
		if key == "" {
			ip := getIP(r)
			key = fmt.Sprintf("ip:%s", ip)
			limit = rl.defaultIP
		}

		// O rate limiting só é aplicado se houver um limite definido
		if limit.MaxRequests > 0 {
			// Incrementa o contador e verifica se o limite foi excedido
			count, err := rl.storage.Increment(r.Context(), key, limit.BlockPeriod)
			if err != nil {
				// Se houver um erro no armazenamento, loga e continua para não bloquear o serviço
				http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
				return
			}

			if count > int64(limit.MaxRequests) {
				// Se o limite foi excedido, retorna 429
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}
		}

		// Se o limite não foi excedido, continua para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// getIP extrai o endereço IP do cliente.
func getIP(r *http.Request) string {
	// Procura por headers de proxy primeiro
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
