package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type addCmd struct {
	id     uint64
	userID string
	amount currency.Currency
}

func parseAddCmd(parts []string) addCmd {
	if len(parts) != 4 {
		abortTx("ADD needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[3])
	abortTxOnError(err, "Could not parse amount in transaction")

	return addCmd{
		id:     id,
		userID: parts[2],
		amount: amount,
	}
}

func (a addCmd) Name() string {
	return fmt.Sprintf("[%d] ADD", a.id)
}

func (a addCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>ADD</command>
		<username>%s</username>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, a.id, a.userID, a.amount.ToFloat(),
	)
}

func (a addCmd) Execute() {
	consoleLog.Warning("Not implemented: ADD")
}
