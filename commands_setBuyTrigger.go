package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

type setBuyTriggerCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSetBuyTriggerCmd(parts []string) setBuyTriggerCmd {
	if len(parts) != 5 {
		abortTx("SET_BUY_TRIGGER needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return setBuyTriggerCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (sbt setBuyTriggerCmd) Name() string {
	return fmt.Sprintf("[%d] SET_BUY_TRIGGER", sbt.id)
}

func (sbt setBuyTriggerCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SET_BUY_TRIGGER</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, sbt.id, sbt.userID,
		sbt.stock, sbt.amount.ToFloat(),
	)

	return types.AuditEvent{
		UserID:    sbt.userID,
		ID:        sbt.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (sbt setBuyTriggerCmd) Execute() {
	autoTxKey := types.AutoTxKey{
		Stock:  sbt.stock,
		UserID: sbt.userID,
		Action: "Buy",
	}
	autoTx, found := workATXStore[autoTxKey]
	if !found {
		// autoTx not set. Fail this trans
		consoleLog.Errorf("Trigger set without Amount for user %s's buy transaction on stock %s. Failing Transaction\n",
			sbt.userID, sbt.stock)
		return
	}

	autoTx.Trigger = sbt.amount

	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	body := autoTx.ToCSV()

	err = ch.Publish(
		"",          // exchange
		autoTxQueue, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/csv",
			Headers: amqp.Table{
				"transType": "autoTxInit",
			},
			Body: []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	fmt.Println("Published successfully")
}
