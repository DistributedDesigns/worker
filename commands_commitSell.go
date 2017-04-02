package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
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

func (cs commitSellCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>COMMIT_SELL</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cs.id, cs.userID,
	)

	return types.AuditEvent{
		UserID:    cs.userID,
		ID:        cs.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (cs commitSellCmd) Execute() {
	abortTxIfNoAccount(cs.userID)

	acct := accountStore[cs.userID]
	pendingSell, err := acct.pendingSells.Pop()
	abortTxOnError(err, cs.Name()+" No pending sells")

	if pendingSell.IsExpired() {
		pendingSell.RollBack()
		abortTx(cs.Name() + " Sell is expired")
	}

	pendingSell.Commit()

	acct.AddSummaryItem("Finished " + cs.Name())
	consoleLog.Notice(" [✔] Finished", cs.Name())
}
