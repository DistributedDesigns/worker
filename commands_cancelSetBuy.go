package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
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
		Action: "Sell",
	}
	delete(workATXStore, autoTxKey)
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	body := autoTxKey.ToCSV()

	err = ch.Publish(
		"",          // exchange
		autoTxQueue, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/csv",
			Headers: amqp.Table{
				"transType": "autoTxKey",
			},
			Body: []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	consoleLog.Debugf("Published aTx %v successfully", autoTxKey)
}
