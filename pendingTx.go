package main

type pendingTxState struct {
	Stock     string `json:"stock"`
	Amount    string `json:"amount"`
	ExpiresAt string `json:"expiresAt"`
}

type pendingTx interface {
	Commit()
	RollBack()
	IsExpired() bool
	GetState() pendingTxState
}
