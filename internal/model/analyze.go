package model

type FraudScoreRequest struct {
	ID          string `json:"id"`
	Transaction struct {
		Amount       float32 `json:"amount"`
		Installments int     `json:"installments"`
		RequestedAt  string  `json:"requested_at"`
	} `json:"transaction"`
	Customer struct {
		Avg_amount     float32  `json:"avg_amount"`
		Tx_count_24h   int      `json:"tx_count_24h"`
		Known_Merchant []string `json:"known_merchant"`
	} `json:"customer"`
	Merchant struct {
		ID        string  `json:"id"`
		MCC       string  `json:"mcc"`
		AvgAmount float32 `json:"avg_amount"`
	} `json:"merchant"`
	Terminal struct {
		IsOnline    bool    `json:"is_online"`
		CardPresent bool    `json:"card_present"`
		KmFromHome  float32 `json:"km_from_home"`
	} `json:"terminal"`
	LastTransaction struct {
		Timestamp     string  `json:"timestamp"`
		KmFromCurrent float32 `json:"km_from_current"`
	} `json:"last_transaction"`
}

type FraudScoreResponse struct {
	Aprrove     bool    `json:"approve"`
	Fraud_score float32 `json:"fraud_score"`
}

type FraudScoreConfig struct {
	MaxAmount            float32 `json:"max_amount"`
	MaxInstallments      int     `json:"max_installments"`
	AmountVsAvgRatio     float32 `json:"amount_vs_avg_ratio"`
	MaxMinutes           int     `json:"max_minutes"`
	MaxKm                float32 `json:"max_km"`
	MaxTxCount24h        int     `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float32 `json:"max_merchant_avg_amount"`
}
