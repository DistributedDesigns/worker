package main

import (
	"fmt"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type quoteCmd struct {
	id     uint64
	userID string
	stock  string
}

func (q quoteCmd) Name() string {
	return fmt.Sprintf("[%d] QUOTE", q.id)
}

func (q quoteCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
  <userCommand>
    <timestamp>%d</timestamp>
    <server>%s</server>
    <transactionNum>%d</transactionNum>
    <command>QUOTE</command>
    <username>%s</username>
    <stockSymbol>%s</stockSymbol>
  </userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, q.id, q.userID, q.stock,
	)
}

func (q quoteCmd) Execute() {
	qr := types.QuoteRequest{
		Stock:      q.stock,
		UserID:     q.userID,
		AllowCache: true,
		ID:         q.id,
	}

	// TODO: actually return a response
	_ = getQuote(qr)
}
