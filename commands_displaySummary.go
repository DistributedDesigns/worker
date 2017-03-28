package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type displaySummaryCmd struct {
	id     uint64
	userID string
}

func parseDisplaySummaryCmd(parts []string) displaySummaryCmd {
	if len(parts) != 3 {
		abortTx("DISPLAY_SUMMARY needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return displaySummaryCmd{
		id:     id,
		userID: parts[2],
	}
}

func (ds displaySummaryCmd) Name() string {
	return fmt.Sprintf("[%d] DISPLAY_SUMMARY", ds.id)
}

func (ds displaySummaryCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>DISPLAY_SUMMARY</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, ds.id, ds.userID,
	)

	return types.AuditEvent{
		UserID:    ds.userID,
		ID:        ds.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (ds displaySummaryCmd) Execute() {
	consoleLog.Warning("Not implemented: DISPLAY_SUMMARY")
}
