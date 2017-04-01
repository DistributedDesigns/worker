package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
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

	return cancelSellCmd{
		id:     id,
		userID: parts[2],
	}
}

func (cs cancelSellCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_SELL", cs.id)
}

func (cs cancelSellCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_SELL</command>
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

func (cs cancelSellCmd) Execute() {
	abortTxIfNoAccount(cs.userID)

	acct := accountStore[cs.userID]
	pendingSell, err := acct.pendingSells.Pop()
	abortTxOnError(err, cs.Name()+" No pending sells")

	pendingSell.RollBack()

	acct.AddSummaryItem("Finished " + cs.Name())
	consoleLog.Notice(" [✔] Finished", cs.Name())
}
