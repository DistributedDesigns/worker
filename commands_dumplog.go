package main

import (
	"fmt"
	"strconv"
	"time"
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

func (dl dumplogCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>DUMPLOG</command>
		<username>%s</username>
		<filename>%s</filename>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, dl.id, dl.userID, dl.filename,
	)
}

func (dl dumplogCmd) Execute() {
	consoleLog.Warning("Not implemented: DUMPLOG")
}
