package scorer

func FraudScore(neighbors []SearchResult) float64 {
	var fraudCount float64
	for _, n := range neighbors {
		if n.Label == "fraud" {
			fraudCount++
		}
	}
	return fraudCount / float64(len(neighbors))
}
