# Use a imagem oficial do Go como base
FROM golang:latest

# Define o diretório de trabalho dentro do container
WORKDIR /app1

# Copia os arquivos de módulo e baixa as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte do projeto para o container
COPY . .

# Expõe as portas para os serviços REST e gRPC
EXPOSE 8080

# Compila o programa Go e gera um binário executável
#RUN go build -o server .
RUN go build -o server .

# Comando para executar a aplicação
CMD ["./server"]
