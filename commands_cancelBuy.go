package main

import (
	"fmt"
	"strconv"
	"time"
)

type cancelBuyCmd struct {
	id     uint64
	userID string
}

func parseCancelBuyCmd(parts []string) cancelBuyCmd {
	if len(parts) != 3 {
		abortTx("CANCEL_BUY needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return cancelBuyCmd{
		id:     id,
		userID: parts[2],
	}
}

func (cb cancelBuyCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_BUY", cb.id)
}

func (cb cancelBuyCmd) GetUserID() string {
	return cb.userID
}

func (cb cancelBuyCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_BUY</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cb.id, cb.userID,
	)
}

func (cb cancelBuyCmd) Execute() {
	consoleLog.Warning("Not implemented: CANCEL_BUY")
}
