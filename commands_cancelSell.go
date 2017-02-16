package main

import (
	"fmt"
	"strconv"
	"time"
)

type cancelSellCmd struct {
	id     uint64
	userID string
}

func parseCancelSellCmd(parts []string) cancelSellCmd {
	if len(parts) != 3 {
		abortTx("CANCEL_SELL needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	// rest of parsing

	return cancelSellCmd{
		id:     id,
		userID: parts[2],
		// rest of parts
	}
}

func (cs cancelSellCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_SELL", cs.id)
}

func (cs cancelSellCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_SELL</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cs.id, cs.userID,
	)
}

func (cs cancelSellCmd) Execute() {
	consoleLog.Warning("Not implemented: CANCEL_SELL")
}
