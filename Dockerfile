FROM golang:1.23-alpine

# Install git (required for go get) and ca-certificates
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Set workdir
WORKDIR /app

# Copy go mod files and download deps
COPY go.mod go.sum ./
RUN go mod download

# Configura o fuso horário para UTC-3 (BRT)
ENV TZ=America/Sao_Paulo
RUN ln -snf /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && echo "America/Sao_Paulo" > /etc/timezone

# Copy source code
COPY . .

# Build the app
RUN go build -o s3sync_minio s3_sync_minio.go

# Command
CMD [ "./s3sync_minio" ]
