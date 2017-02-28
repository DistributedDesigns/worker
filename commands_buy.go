package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type buyCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseBuyCmd(parts []string) buyCmd {
	if len(parts) != 5 {
		abortTx("BUY needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return buyCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (b buyCmd) Name() string {
	return fmt.Sprintf("[%d] BUY", b.id)
}

func (b buyCmd) GetUserID() string {
	return b.userID
}

func (b buyCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>BUY</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, b.id, b.userID, b.stock, b.amount.ToFloat(),
	)
}

func (b buyCmd) Execute() {
	consoleLog.Warning("Not implemented: BUY")
}
