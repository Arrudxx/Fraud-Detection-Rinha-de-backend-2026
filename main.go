//go:build !windows

package main

import (
	"fmt"
	"fraud-detection/internal/api"
	"fraud-detection/internal/scorer"
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

const (
	jsonPath  = "/app/data/references.json"
	binPath   = "/data/shared/dataset.bin"
	sharedDir = "/data/shared"
)

func main() {
	// garante que o dataset.bin existe no volume compartilhado
	if err := scorer.EnsureDatasetBin(jsonPath, binPath, sharedDir); err != nil {
		log.Fatal("erro ao preparar dataset:", err)
	}

	// carrega o índice LSH via mmap
	log.Println("carregando índice via mmap...")
	index, err := scorer.LoadMmapIndex(binPath)
	if err != nil {
		log.Fatal("erro ao carregar índice:", err)
	}
	log.Println("índice pronto")

	runtime.GC()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	fmt.Printf("memória usada: %v MB\n", mem.Alloc/1024/1024)
	log.Printf("cores disponíveis: %d\n", runtime.NumCPU())

	router := gin.Default()
	router.GET("/ready", api.Ready)
	router.POST("/fraud-score", api.FraudScore(index))
	router.Run(":8080")
}
