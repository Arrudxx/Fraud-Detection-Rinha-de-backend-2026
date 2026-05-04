package api

import (
	"fmt"
	"fraud-detection/internal/model"
	"fraud-detection/internal/scorer"

	"github.com/gin-gonic/gin"
)

func FraudScore(c *gin.Context) {

	// Usa modelo definido para receber os dados da requisição
	var req model.FraudScoreRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	riskScore := scorer.CalculateRisk(req.Transaction.Amount)

	fmt.Printf("Calculated risk score: %f\n", riskScore)

	c.JSON(200, model.FraudScoreResponse{
		Aprrove:     true,
		Fraud_score: riskScore,
	})
}
