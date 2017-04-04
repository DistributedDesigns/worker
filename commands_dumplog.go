package main

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

type dumplogCmd struct {
	id       uint64
	userID   string
	filename string
}

func parseDumplogCmd(parts []string) dumplogCmd {
	if len(parts) < 3 {
		abortParse("DUMPLOG needs at least 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortParseOnError(err, "Could not parse ID")

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

	return dumplogCmd{id, userID, safeFileName(filename)}
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
	abortTxIfNoAccount(dl.userID)

	dlr := types.DumplogRequest{
		UserID:   dl.userID,
		Filename: dl.filename,
	}

	// Optimistically send request. It's up to the user to retrieve the file~
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Publish(
		"",       // exchange
		dumplogQ, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/csv",
			Body:        []byte(dlr.ToCSV()),
		})
	failOnError(err, "Failed to publish a message")

	consoleLog.Debug("Dumplog requested as", dlr.Filename)

	acct := accountStore[dl.userID]
	acct.PushEvent(fmt.Sprintf("Wrote dumplog to %s\nContact an admin to retrieve your file", dlr.Filename))
	acct.AddSummaryItem("Finished " + dl.Name())

	consoleLog.Notice(" [✔] Finished", dl.Name())
}

// Compile the file sanitization regexps once and only once
var (
	separators = regexp.MustCompile(`[ &_=+:]`)
	legal      = regexp.MustCompile(`[^[:alnum:]-.]`)
)

// Convert to a string that's suitable for use as a filename.
// Lifted from asaskevich/govalidator
func safeFileName(str string) string {
	name := strings.ToLower(str)
	name = path.Clean(path.Base(name)) // "./foo/bar" -> "bar"
	name = strings.Trim(name, " ")
	name = separators.ReplaceAllString(name, "-")
	name = legal.ReplaceAllString(name, "")
	name = strings.Replace(name, "--", "-", -1)
	return name
}
