package main

import (
	"fmt"
	"strconv"
	"time"
)

type commitSellCmd struct {
	id     uint64
	userID string
}

func parseCommitSellCmd(parts []string) commitSellCmd {
	if len(parts) != 3 {
		abortTx("COMMIT_SELL needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return commitSellCmd{
		id:     id,
		userID: parts[2],
	}
}

func (cs commitSellCmd) Name() string {
	return fmt.Sprintf("[%d] COMMIT_SELL", cs.id)
}

func (cs commitSellCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>COMMIT_SELL</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cs.id, cs.userID,
	)
}

func (cs commitSellCmd) Execute() {
	consoleLog.Warning("Not implemented: COMMIT_SELL")
}
