package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type sellCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSellCmd(parts []string) sellCmd {
	if len(parts) != 5 {
		abortTx("SELL needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	// rest of parsing

	return sellCmd{
		id:     id,
		userID: parts[2],
		// rest of parts
	}
}

func (s sellCmd) Name() string {
	return fmt.Sprintf("[%d] SELL", s.id)
}

func (s sellCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SELL</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, s.id, s.userID,
		s.stock, s.amount.ToFloat(),
	)
}

func (s sellCmd) Execute() {
	consoleLog.Warning("Not implemented: SELL")
}
