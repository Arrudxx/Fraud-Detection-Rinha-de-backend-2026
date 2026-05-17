package scorer

import (
	"fraud-detection/internal/model"
	"time"
)

func boolToRisk(b bool) float32 {
	if b {
		return 1.0
	}
	return 0.0
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func clamp(dividendo, divisor float32) float32 {
	if divisor == 0 {
		return 0
	}
	result := dividendo / divisor
	if result > 1.0 {
		return 1.0
	}
	if result < 0.0 {
		return 0.0
	}
	return result
}

func calculateHourOfDay(requestedAt string) float32 {
	t, err := time.Parse(time.RFC3339, requestedAt)
	if err != nil {
		return 0.0
	}
	return float32(t.UTC().Hour()) / 23.0
}

func calculateDayOfWeek(requestedAt string) float32 {
	t, err := time.Parse(time.RFC3339, requestedAt)
	if err != nil {
		return 0.0
	}
	dayOfWeek := (int(t.UTC().Weekday()) + 6) % 7
	return float32(dayOfWeek) / 6.0
}

func calculateMinutesSinceLastTransaction(requestedAt string, lastTransaction string, maxMinutes float32) float32 {
	if lastTransaction == "" {
		return -1.0
	}

	t1, err1 := time.Parse(time.RFC3339, requestedAt)
	t2, err2 := time.Parse(time.RFC3339, lastTransaction)
	if err1 != nil || err2 != nil {
		return -1.0
	}

	minutes := float32(t1.Sub(t2).Minutes())
	result := minutes / maxMinutes

	if result < 0 {
		return 0
	}
	if result > 1.0 {
		return 1.0
	}
	return result
}

func TransactionRisk(req model.FraudScoreRequest) []float32 {
	return []float32{
		clamp(req.Transaction.Amount, MaxAmount),
		clamp(float32(req.Transaction.Installments), MaxInstallments),
		clamp(clamp(req.Transaction.Amount, req.Customer.Avg_amount), AmountVsAvgRatio),
		calculateHourOfDay(req.Transaction.RequestedAt),
		calculateDayOfWeek(req.Transaction.RequestedAt),
		calculateMinutesSinceLastTransaction(req.Transaction.RequestedAt, req.LastTransaction.Timestamp, MaxMinutes),
		clamp(req.LastTransaction.KmFromCurrent, MaxKm),
		clamp(req.Terminal.KmFromHome, MaxKm),
		clamp(float32(req.Customer.Tx_count_24h), MaxTxCount24h),
		boolToRisk(req.Terminal.IsOnline),
		boolToRisk(req.Terminal.CardPresent),
		boolToRisk(contains(req.Customer.Known_Merchant, req.Merchant.ID)),
		mccRiskScores[req.Merchant.MCC],
		clamp(req.Merchant.AvgAmount, MaxMerchantAvgAmount),
	}
}
