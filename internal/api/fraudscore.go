package api

import (
	"fraud-detection/internal/model"
	"fraud-detection/internal/scorer"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FraudScore(index *scorer.HNSWIndex) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Usa modelo definido para receber os dados da requisição
		var req model.FraudScoreRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// // Transforma os dados em um vetor de 14 dimensões
		// riskScore := scorer.TransactionRisk(req)

		// fmt.Println("Vetor de risco:", riskScore)

		// // Busca os 5 vizinhos mais próximos usando o vetor de risco
		// neighbors := scorer.KNNSearch(riskScore, refs, 5)

		// fmt.Println("Vizinhos mais próximos:", neighbors)

		// // Calcula o score
		// score := scorer.FraudScore(neighbors)

		// fmt.Printf("Score de fraude calculado: %.4f\n", score)

		vector := scorer.TransactionRisk(req)

		// fmt.Println(vector)

		neighbors := index.Search(vector, 5) // ← usa o índice agora
		score := scorer.FraudScore(neighbors)

		c.JSON(http.StatusOK, gin.H{
			"approved":    score < 0.6,
			"fraud_score": score,
		})
	}
}
