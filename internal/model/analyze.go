package model

type FraudScoreRequest struct {
	ID          string `json:"id"`
	Transaction struct {
		Amount       float64 `json:"amount"`
		Installments int     `json:"installments"`
		RequestedAt  string  `json:"requested_at"`
	} `json:"transaction"`
	Customer struct {
		Avg_amount     float64  `json:"avg_amount"`
		Tx_count_24h   int      `json:"tx_count_24h"`
		Known_Merchant []string `json:"known_merchant"`
	} `json:"customer"`
	Merchant struct {
		ID        string  `json:"id"`
		MCC       string  `json:"mcc"`
		AvgAmount float64 `json:"avg_amount"`
	} `json:"merchant"`
	Terminal struct {
		IsOnline    bool    `json:"is_online"`
		CardPresent bool    `json:"card_present"`
		KmFromHome  float64 `json:"km_from_home"`
	} `json:"terminal"`
	LastTransaction struct {
		Timestamp     string  `json:"timestamp"`
		KmFromCurrent float64 `json:"km_from_current"`
	} `json:"last_transaction"`
}

type FraudScoreResponse struct {
	Aprrove     bool    `json:"approve"`
	Fraud_score float64 `json:"fraud_score"`
}
