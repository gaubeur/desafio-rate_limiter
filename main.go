package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gaubeur/desafio-rate_limiter/internal/limiter"
	"github.com/gaubeur/desafio-rate_limiter/internal/storage"
	"github.com/go-redis/redis/v8"

	"github.com/joho/godotenv"
)

// homeHandler lida com a lógica de negócio principal da sua aplicação.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "FullCycle: Desafio Rate Limiter")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar arquivo .env: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tokenALimit, err := strconv.Atoi(os.Getenv("TOKEN_A_LIMIT"))
	if err != nil {
		log.Fatalf("Erro ao converter TOKEN_A_LIMIT para inteiro: %v\n", err)
	}

	tokenBLimit, err := strconv.Atoi(os.Getenv("TOKEN_B_LIMIT"))
	if err != nil {
		log.Fatalf("Erro ao converter TOKEN_B_LIMIT para inteiro: %v\n", err)
	}

	tokenADuration, err := time.ParseDuration(os.Getenv("TOKEN_A_BLOCK_DURATION"))
	if err != nil {
		log.Fatalf("Erro ao converter TOKEN_A_BLOCK_DURATION para inteiro: %v\n", err)
	}

	tokenBDuration, err := time.ParseDuration(os.Getenv("TOKEN_B_BLOCK_DURATION"))
	if err != nil {
		log.Fatalf("Erro ao converter TOKEN_B_BLOCK_DURATION para inteiro: %v\n", err)
	}

	tokenLimits := map[string]limiter.Limit{
		"token-A": {MaxRequests: tokenALimit, BlockPeriod: tokenADuration},
		"token-B": {MaxRequests: tokenBLimit, BlockPeriod: tokenBDuration},
	}

	var defaultIP limiter.Limit
	defaultIP.MaxRequests, err = strconv.Atoi(os.Getenv("IP_LIMIT"))
	if err != nil {
		log.Fatalf("Erro ao converter IP_LIMIT para inteiro: %v\n", err)
	}
	defaultIP.BlockPeriod, err = time.ParseDuration(os.Getenv("IP_BLOCK_DURATION"))
	if err != nil {
		log.Fatalf("Erro ao converter IP_BLOCK_DURATION para inteiro: %v\n", err)
	}

	var s storage.Storage

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Printf("Usando Redis. Conectando em %s...", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Não foi possível pingar o Redis: %v", err)
	}
	log.Println("Conexão com o Redis estabelecida com sucesso!")
	s = storage.NewRedisStorage(rdb)

	rl := limiter.NewRateLimiter(s, defaultIP.MaxRequests, defaultIP.BlockPeriod, tokenLimits)

	handler := http.HandlerFunc(HomeHandler)
	protectedHandler := rl.Middleware(handler)

	mux := http.NewServeMux()
	mux.Handle("/", protectedHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor iniciado em http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Não foi possível iniciar o servidor: %v", err)
		}
	}()

	<-done
	log.Println("Servidor recebendo sinal de shutdown...")
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Shutdown do servidor falhou: %v", err)
	}

	log.Println("Servidor encerrado com sucesso.")
}
