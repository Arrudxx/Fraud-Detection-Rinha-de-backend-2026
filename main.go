package main

import (
	"encoding/json"
	"fraud-detection/internal/api"
	"fraud-detection/internal/scorer"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	// 1. abre o arquivo
	file, err := os.Open("data/references.json")
	if err != nil {
		log.Fatal("erro ao abrir references.json:", err)
	}
	defer file.Close()

	// 2. lê e converte para []Reference
	var refs []scorer.Reference
	if err := json.NewDecoder(file).Decode(&refs); err != nil {
		log.Fatal("erro ao decodificar references.json:", err)
	}

	log.Printf("%d referências carregadas na memória\n", len(refs))

	// constrói o índice HNSW (pode levar 30-60s com 3M vetores)
	log.Println("construindo índice HNSW...")
	index, err := scorer.BuildIndex(refs)
	if err != nil {
		log.Fatal("erro ao construir índice:", err)
	}
	log.Println("índice pronto")

	// Gin router
	router := gin.Default()
	// Endpoint health check
	router.GET("/ready", api.Ready)
	// Endpoint para calcular o score de fraude
	router.POST("/fraud-score", api.FraudScore(index))
	router.Run(":8080")

}
