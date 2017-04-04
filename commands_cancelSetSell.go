package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type cancelSetSellCmd struct {
	id     uint64
	userID string
	stock  string
}

func parseCancelSetSellCmd(parts []string) cancelSetSellCmd {
	if len(parts) != 4 {
		abortParse("CANCEL_SET_SELL needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortParseOnError(err, "Could not parse ID")

	return cancelSetSellCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
	}
}

func (css cancelSetSellCmd) Name() string {
	return fmt.Sprintf("[%d] CANCEL_SET_SELL", css.id)
}

func (css cancelSetSellCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
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

	return types.AuditEvent{
		UserID:    css.userID,
		ID:        css.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (css cancelSetSellCmd) Execute() {
	autoTxKey := types.AutoTxKey{
		Stock:  css.stock,
		UserID: css.userID,
		Action: "Sell",
	}
	delete(workATXStore, autoTxKey)
	autoTxCancelChan <- autoTxKey
	consoleLog.Debugf("Published aTx %v successfully", autoTxKey)

	acct := accountStore[css.userID]
	acct.AddSummaryItem("Finished " + css.Name())
	consoleLog.Notice(" [✔] Finished", css.Name())
}
