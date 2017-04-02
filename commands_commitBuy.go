package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
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

func (cb commitBuyCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>COMMIT_BUY</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cb.id, cb.userID,
	)

	return types.AuditEvent{
		UserID:    cb.userID,
		ID:        cb.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (cb commitBuyCmd) Execute() {
	abortTxIfNoAccount(cb.userID)

	acct := accountStore[cb.userID]
	pendingBuy, err := acct.pendingBuys.Pop()
	abortTxOnError(err, cb.Name()+" No pending buys")

	if pendingBuy.IsExpired() {
		pendingBuy.RollBack()
		abortTx(cb.Name() + " Buy expired")
	}

	pendingBuy.Commit()

	acct.AddSummaryItem("Finished " + cb.Name())
	consoleLog.Notice(" [✔] Finished", cb.Name())
}
