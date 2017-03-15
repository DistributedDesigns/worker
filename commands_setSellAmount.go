package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type setSellAmountCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSetSellAmountCmd(parts []string) setSellAmountCmd {
	if len(parts) != 5 {
		abortTx("SET_SELL_AMOUNT needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return setSellAmountCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (ssa setSellAmountCmd) Name() string {
	return fmt.Sprintf("[%d] SET_SELL_AMOUNT", ssa.id)
}

func (ssa setSellAmountCmd) GetUserID() string {
	return ssa.userID
}

func (ssa setSellAmountCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SET_SELL_AMOUNT</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, ssa.id, ssa.userID,
		ssa.stock, ssa.amount.ToFloat(),
	)
}

func (ssa setSellAmountCmd) Execute() {
	consoleLog.Warning("Not implemented: SET_SELL_AMOUNT")
}
