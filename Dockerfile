#build
# usa a imagem oficial do Go para compilar

FROM golang:1.26-alpine AS builder

WORKDIR /app

# copia os arquivos de dependência primeiro
COPY go.mod go.sum ./
RUN go mod download

# copia o resto do código
COPY . .

# compila para linux/amd64
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o fraud-detection ./main.go

#runtime
FROM alpine:3.19

WORKDIR /app

# copia só o binário compilado da etapa anterior
COPY --from=builder /app/fraud-detection .

# copia o dataset
COPY data/ ./data/

#porta
EXPOSE 8080

#Executa o binário da API
CMD ["./fraud-detection"]