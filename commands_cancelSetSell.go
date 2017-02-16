package main

import (
	"fmt"
	"strconv"
	"time"
)

type cancelSetSellCmd struct {
	id     uint64
	userID string
	stock  string
}

func parseCancelSetSellCmd(parts []string) cancelBuyCmd {
	if len(parts) != 4 {
		abortTx("CANCEL_SET_SELL needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	// rest of parsing

	return cancelBuyCmd{
		id:     id,
		userID: parts[2],
		// rest of parts
	}
}

func (css cancelSetSellCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_SET_SELL", css.id)
}

func (css cancelSetSellCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_SET_SELL</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, css.id, css.userID, css.stock,
	)
}

func (css cancelSetSellCmd) Execute() {
	consoleLog.Warning("Not implemented: CANCEL_SET_SELL")
}
