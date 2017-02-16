package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type setBuyAmountCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSetBuyAmountCmd(parts []string) setBuyAmountCmd {
	if len(parts) != 5 {
		abortTx("SET_BUY_AMOUNT needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return setBuyAmountCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (sba setBuyAmountCmd) Name() string {
	return fmt.Sprintf("[%d] SET_BUY_AMOUNT", sba.id)
}

func (sba setBuyAmountCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SET_BUY_AMOUNT</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, sba.id, sba.userID,
		sba.stock, sba.amount.ToFloat(),
	)
}

func (sba setBuyAmountCmd) Execute() {
	consoleLog.Warning("Not implemented: SET_BUY_AMOUNT")
}
