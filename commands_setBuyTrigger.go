package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type setBuyTriggerCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSetBuyTriggerCmd(parts []string) setBuyTriggerCmd {
	if len(parts) != 5 {
		abortTx("SET_BUY_TRIGGER needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	// rest of parsing

	return setBuyTriggerCmd{
		id:     id,
		userID: parts[2],
		// rest of parts
	}
}

func (sbt setBuyTriggerCmd) Name() string {
	return fmt.Sprintf("[%d] SET_BUY_TRIGGER", sbt.id)
}

func (sbt setBuyTriggerCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SET_BUY_TRIGGER</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, sbt.id, sbt.userID,
		sbt.stock, sbt.amount.ToFloat(),
	)
}

func (sbt setBuyTriggerCmd) Execute() {
	consoleLog.Warning("Not implemented: SET_BUY_TRIGGER")
}
