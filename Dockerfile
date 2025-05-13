FROM golang:1.23-alpine

# Install git (required for go get) and ca-certificates
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Set workdir
WORKDIR /app

# Copy go mod files and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the app
RUN go build -o s3sync_minio s3_sync_minio.go

# Usando uma imagem ARM do Debian para a execução
FROM debian:bookworm

# Atualiza o sistema e instala o cliente MySQL
RUN apt-get update && \
    apt-get install -y cron tzdata && \
    rm -rf /var/lib/apt/lists/*

# Configura o fuso horário para UTC-3 (BRT)
ENV TZ=America/Sao_Paulo
RUN ln -snf /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && echo "America/Sao_Paulo" > /etc/timezone

# Configura o diretório de trabalho no novo container
WORKDIR /app

# Copia o executável da fase de compilação para esta fase
COPY --from=builder /app/s3_sync_minio .

# Copia o arquivo .env
COPY .env .

# Adiciona uma crontab para executar o script uma vez por dia
RUN echo "0 3 * * * /app/s3_sync_minio >> /var/log/s3sync.log 2>&1" > /etc/cron.d/s3sync-cron && \
    chmod 0644 /etc/cron.d/s3sync-cron && \
    crontab /etc/cron.d/s3sync-cron

# Command
# CMD [ "./s3sync_minio" ]

# Comando para iniciar a aplicação
# Inicia o cron e o shell
CMD ["cron", "-f"]
