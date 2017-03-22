package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type dumplogCmd struct {
	id       uint64
	userID   string
	filename string
}

func parseDumplogCmd(parts []string) dumplogCmd {
	if len(parts) < 3 {
		abortTx("DUMPLOG needs at least 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	// Dumplog is overloaded as
	// 1) DUMPLOG,$filename
	// 2) DUMPLOG,$userID,$filename
	// The first case is an admin (global) dump
	var userID, filename string
	if len(parts) == 3 {
		userID = "admin"
		filename = parts[2]
	} else {
		userID = parts[2]
		filename = parts[3]
	}

	return dumplogCmd{id, userID, filename}
}

func (dl dumplogCmd) Name() string {
	return fmt.Sprintf("[%d] DUMPLOG", dl.id)
}

func (dl dumplogCmd) ToAuditEvent() types.AuditEvent {
	// Workload file validator will fail if it sees the injected "admin" account
	// in the workload file. We'll make it disappear here but carry it through
	// the rest of the system.
	var userNameField string
	if dl.userID != "admin" {
		userNameField = fmt.Sprintf("\n\t\t<username>%s</username>", dl.userID)
	}

	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>DUMPLOG</command>%s
		<filename>%s</filename>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, dl.id, userNameField, dl.filename,
	)

	return types.AuditEvent{
		UserID:    dl.userID,
		ID:        dl.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (dl dumplogCmd) Execute() {
	consoleLog.Warning("Not implemented: DUMPLOG")
}
