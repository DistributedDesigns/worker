package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type quoteCmd struct {
	id     uint64
	userID string
	stock  string
}

func parseQuoteCmd(parts []string) quoteCmd {
	if len(parts) != 4 {
		abortTx("QUOTE needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return quoteCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
	}
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

	consoleLog.Notice(" [✔] Finished", q.Name())
}