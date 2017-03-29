package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
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

func (sba setBuyAmountCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
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

	return types.AuditEvent{
		UserID:    sba.userID,
		ID:        sba.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (sba setBuyAmountCmd) Execute() {
	autoTxKey := types.AutoTxKey{
		Stock:  sba.stock,
		UserID: sba.userID,
		Action: "Buy",
	}
	_, found := workATXStore[autoTxKey]
	if found {
		// autoTx already exists, we'll need to cancel it.
		// TODO

	}

	workATXStore[autoTxKey] = types.AutoTxInit{
		AutoTxKey: autoTxKey,
		Amount:    sba.amount,
		WorkerID:  *workerNum,
	}
	//fmt.Println(workATXStore)
	// consoleLog.Warning("Not implemented: SET_BUY_AMOUNT")
}
