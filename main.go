package main

import (
	"fraud-detection/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	// Endpoint health check
	router.GET("/ready", api.Ready)

	// Endpoint para calcular o score de fraude
	router.POST("/fraud-score", api.FraudScore)

	router.Run(":9999")

}
