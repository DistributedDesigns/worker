package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

var (
	loggableCmds = make(chan command)
	logChannel   *amqp.Channel
	cmdStartTime time.Time
	elapsedTime  int64 // preallocate space
)

func txWorker(unprocessedTxs <-chan string) {
	// create a single channel for logging so we don't flood RMQ
	// with new channels for each log item.
	var err error
	logChannel, err = rmqConn.Channel()
	failOnError(err, "Failed to open a logging channel")
	defer logChannel.Close()

	go sendCmdToAudit()

	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished processing transactions")
			cleanUpTxs(unprocessedTxs)
			return
		default:
			processTxs(unprocessedTxs)
		}
	}
}

func processTxs(unprocessedTxs <-chan string) {
	// If a TX is aborted anwhere in its processing it will bubble up
	// to here and catchAbortedTx() will run before processTxs() closes.
	// Control returns to txWorker() where the select{} repeats and the
	// next transaction is grabbed.
	defer catchAbortedTx()

	tx := <-unprocessedTxs
	cmdStartTime = time.Now()
	cmd := parseCommand(tx)

	if !*noAudit {
		loggableCmds <- cmd
	}

	cmd.Execute()

	// Only log runtime if command ran to completion
	elapsedTime = time.Since(cmdStartTime).Nanoseconds()

	// "[123] BUY" -> "BUY"
	cmdType := strings.Split(cmd.Name(), " ")[1]
	fmt.Fprintf(os.Stderr, "%s,%d\n", cmdType, elapsedTime)
}

func abortTx(msg string) {
	panic(msg)
}

func abortTxOnError(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}

func catchAbortedTx() {
	if r := recover(); r != nil {
		consoleLog.Error(" [x] Aborted transaction:", r)
	}
}

func cleanUpTxs(unprocessedTxs <-chan string) {
	// TODO: Put these back in redis? Just warn for now.
	if len(unprocessedTxs) > 0 {
		for i := 0; i < len(unprocessedTxs); i++ {
			consoleLog.Warning("Unprocessed transaction", <-unprocessedTxs)
		}
	}
}

func sendCmdToAudit() {
	if *noAudit {
		consoleLog.Warning("Not sending to audit log")
		return
	}

	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished sending commands to log")
			cleanUpCmdLog()
			return
		case cmd := <-loggableCmds:
			header := amqp.Table{
				"name":      cmd.Name(),
				"serviceID": redisBaseKey,
				"userID":    cmd.GetUserID(),
			}

			err := logChannel.Publish(
				"",          // exchange
				auditEventQ, // routing key
				false,       // mandatory
				false,       // immediate
				amqp.Publishing{
					Headers:     header,
					ContentType: "text/plain",
					Body:        []byte(cmd.ToAuditEntry()),
				})
			failOnError(err, "Failed to publish a message")

			consoleLog.Debug("Sent to audit:", cmd.Name())
		}
	}
}

func cleanUpCmdLog() {
	if len(loggableCmds) > 0 {
		for i := 0; i < len(loggableCmds); i++ {
			cmd := <-loggableCmds
			consoleLog.Warning("Unlogged command", cmd.Name())
		}
	}
}
