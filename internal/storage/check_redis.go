// Este programa Go testa a conexão com um servidor Redis.
package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func checkRedis() {
	// Cria um novo cliente Redis.
	// O endereço 127.0.0.1:6379 é o padrão.
	// Se o seu Redis estiver noutra porta, altere aqui.
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Usa um timeout para a conexão.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tenta fazer um PING para o servidor Redis.
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		// Se houver um erro, imprima a mensagem.
		// Isso pode ser uma falha na conexão, permissões, ou firewall.
		log.Fatalf("Falha ao conectar-se ao Redis: %v", err)
	}

	// Se a conexão for bem-sucedida, imprima a resposta PONG.
	fmt.Println("Conexão bem-sucedida ao Redis!")
	fmt.Printf("Resposta do PING: %s\n", pong)
}
