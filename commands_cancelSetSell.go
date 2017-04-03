package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

type cancelSetSellCmd struct {
	id     uint64
	userID string
	stock  string
}

func parseCancelSetSellCmd(parts []string) cancelSetSellCmd {
	if len(parts) != 4 {
		abortTx("CANCEL_SET_SELL needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

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
	//Copypasta between cancelSetBuy and cancelSetSell
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
