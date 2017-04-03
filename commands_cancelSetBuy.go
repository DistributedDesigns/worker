package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type cancelSetBuyCmd struct {
	id     uint64
	userID string
	stock  string
}

func parseCancelSetBuyCmd(parts []string) cancelSetBuyCmd {
	if len(parts) != 4 {
		abortTx("CANCEL_SET_BUY needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return cancelSetBuyCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
	}
}

func (csb cancelSetBuyCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_SET_BUY", csb.id)
}

func (csb cancelSetBuyCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>CANCEL_SET_BUY</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, csb.id, csb.userID, csb.stock,
	)

	return types.AuditEvent{
		UserID:    csb.userID,
		ID:        csb.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (csb cancelSetBuyCmd) Execute() {
	autoTxKey := types.AutoTxKey{
		Stock:  csb.stock,
		UserID: csb.userID,
		Action: "Buy",
	}
	delete(workATXStore, autoTxKey)
	autoTxCancelChan <- autoTxKey
	consoleLog.Debugf("Published aTx %v successfully", autoTxKey)
}
