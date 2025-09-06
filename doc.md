Documentação : Desafio Rate Limiter em Go

A idéia principal da implementação de um rate limiter na arquitetura de um servidor web é a de controlar a quantidade de requisições que um cliente pode fazer a um servidor em um determinado período de tempo. No desafio proposto, estamos implementando a limitação com base no IP ou em Token sendo que no caso de token, podemos ter vários tipos com vários cenários de autorização. Para este caso em especifico, definimos o token-A e o token-B e como podemos ver, suas definições são configuraveis em um arquivo .env ou por variáveis de ambiente

Para persistência e gerenciamento do estado do rate limiter, o projeto adotou o uso do nosql Redis. O Redis é muito bom em operações atômicas de incrementação e em definir expirações de chaves, permitindo que o rate limiter funcione de maneira performática. No desafio, usei um pipeline para combinar INCR e EXPIRE. O Redis garante que, dentro desse pipeline, os comandos são executados na ordem que foram enviados. Para garantir que um conjunto de comandos seja executado de forma totalmente atômica, onde falha um, falha todos (rollback), implementei o Pipeline do go-redis que tem a proposta de encapsular o comando MULTI/EXEC para um comportamento mais seguro.

Estratégia Configurável: A lógica de persistência é desacoplada, permitindo a troca fácil do Redis por outro banco de dados ou mecanismo de armazenamento.

As razões para se implementar um rate limiter são inúmeras mas podemos destacar :
Prevensão a ataques
Garantia da qualidade do serviço
Controle de custos
Monetização e gerenciamento de plano

Existe várias maneiras de implementar um rate limiter como um middleware para um servidor web mas no nosso desafio usamos Fixed Window Counter (Contador de Janela Fixa) utilizando o Redis para armazenar os contadores e o tempo de expiração de cada chave (IP ou Token). A chave é identificada usando o endereço IP do remetente ou o API_KEY do cabeçalho da requisição. O limite é aplicado usando o Redis que armazena um contador para cada par (identificador, janela de tempo). Cada vez que uma requisição com um determinado IP ou Token é feita, o contador é incrementado. O gerenciamento da expiração ocorre com a definição do EXPIRE na chave inserida no redis e garante que o contador tenha um tempo de vida definido. Uma vez que esse tempo expira, a chave é removida do Redis, e o contador começa novamente do zero para uma nova janela.

Testando a Aplicação
A aplicação estará acessível em http://localhost:8080.

Podemos testar o rate limiter com ferramentas como curl ou Postman.

# Teste com IP

for i in $(seq 1 6); do
    curl -s -w "HTTP Code: %{http_code}" \
         -H "X-Forwarded-For: 192.168.1.1" \
         "http://localhost:8080";
    echo "";
    sleep 0.05;
done

for i in {1..15}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080; done

Testando a Limitação por Token:

A requisição deve incluir o header API_KEY com o valor do token configurado.

# Teste com o TOKEN_A
for i in {1..25}; do curl -s -o /dev/null -w "%{http_code}\n" -H "API_KEY: token-a-secreto" http://localhost:8080; done

# Teste com o TOKEN_B
for i in {1..55}; do curl -s -o /dev/null -w "%{http_code}\n" -H "API_KEY: token-b-secreto" http://localhost:8080; done
