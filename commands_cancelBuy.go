package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
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

func (cb cancelBuyCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_BUY</command>
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

func (cb cancelBuyCmd) Execute() {
	abortTxIfNoAccount(cb.userID)

	// Pop buy from user's pendingBuys stack
	acct := accountStore[cb.userID]
	pendingBuy, err := acct.pendingBuys.pop()
	abortTxOnError(err, "User has no pending buys")

	pendingBuy.RollBack()

	consoleLog.Notice(" [✔] Finished", cb.Name())
}
