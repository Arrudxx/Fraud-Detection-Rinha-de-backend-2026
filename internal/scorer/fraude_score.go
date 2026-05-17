package scorer

func FraudScore(neighbors []SearchResult) float32 {
	var fraudCount float32
	for _, n := range neighbors {
		if n.Label == "fraud" {
			fraudCount++
		}
	}
	return fraudCount / float32(len(neighbors))
}
