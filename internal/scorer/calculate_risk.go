package scorer

import (
	"fraud-detection/internal/model"
	"math"
	"time"
)

func boolToRisk(b bool) float64 {
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

func TransactionRisk(req model.FraudScoreRequest) []float64 {
	return []float64{
		clamp(req.Transaction.Amount, MaxAmount),                                                                     //amount
		clamp(float64(req.Transaction.Installments), MaxInstallments),                                                //installments
		clamp(clamp(req.Transaction.Amount, req.Customer.Avg_amount), AmountVsAvgRatio),                              //amount_vs_avg
		calculateHourOfDay(req.Transaction.RequestedAt),                                                              //hour_of_day
		calculateDayOfWeek(req.Transaction.RequestedAt),                                                              //day_of_week
		calculateMinutesSinceLastTransaction(req.Transaction.RequestedAt, req.LastTransaction.Timestamp, MaxMinutes), //minutes_since_last_tx
		clamp(req.LastTransaction.KmFromCurrent, MaxKm),                                                              //km_from_last_tx
		clamp(req.Terminal.KmFromHome, MaxKm),                                                                        //km_from_home
		clamp(float64(req.Customer.Tx_count_24h), MaxTxCount24h),                                                     //tx_count_24h
		boolToRisk(req.Terminal.IsOnline),                                                                            //is_online
		boolToRisk(req.Terminal.CardPresent),                                                                         //card_present
		boolToRisk(contains(req.Customer.Known_Merchant, req.Merchant.ID)),                                           //unknown_merchant
		mccRiskScores[req.Merchant.MCC],                                                                              //mcc_risk
		clamp(req.Merchant.AvgAmount, MaxMerchantAvgAmount),                                                          //merchant_avg_amount
	}

}

func normalizeDecimal(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func clamp(dividendo, divisor float64) float64 {
	result := dividendo / divisor
	return normalizeDecimal(result)
}

func calculateHourOfDay(requestedAt string) float64 {
	t, err := time.Parse(time.RFC3339, requestedAt)
	if err != nil {
		return 0.0
	}
	hour := float64(t.UTC().Hour())
	return normalizeDecimal((hour / 23.0))
}

func calculateDayOfWeek(requestedAt string) float64 {
	t, err := time.Parse(time.RFC3339, requestedAt)
	if err != nil {
		return 0.0
	}
	// time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
	// Precisamos: Monday=0, Tuesday=1, ..., Sunday=6
	dayOfWeek := (int(t.UTC().Weekday()) + 6) % 7
	return normalizeDecimal((float64(dayOfWeek) / 6.0))

}

func calculateMinutesSinceLastTransaction(requestedAt string, lastTransaction string, maxMinutes float64) float64 {
	if lastTransaction == "" {
		return -1.0
	}

	t1, err1 := time.Parse(time.RFC3339, requestedAt)
	t2, err2 := time.Parse(time.RFC3339, lastTransaction)
	if err1 != nil || err2 != nil {
		return -1.0
	}

	minutes := t1.Sub(t2).Minutes()
	result := minutes / maxMinutes

	// Limitar entre 0 e 1
	return math.Min(math.Max(result, 0), 1.0)
}
