package main

import (
	"fmt"
	"strconv"
	"time"
)

type commitBuyCmd struct {
	id     uint64
	userID string
}

func parseCommitBuyCmd(parts []string) commitBuyCmd {
	if len(parts) != 3 {
		abortTx("COMMIT_BUY needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return commitBuyCmd{
		id:     id,
		userID: parts[2],
	}
}

func (cb commitBuyCmd) Name() string {
	return fmt.Sprintf("[%d] COMMIT_BUY", cb.id)
}

func (cb commitBuyCmd) GetUserID() string {
	return cb.userID
}

func (cb commitBuyCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>COMMIT_BUY</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cb.id, cb.userID,
	)
}

func (cb commitBuyCmd) Execute() {
	consoleLog.Warning("Not implemented: COMMIT_BUY")
}
