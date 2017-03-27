package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

type setSellTriggerCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
}

func parseSetSellTriggerCmd(parts []string) setSellTriggerCmd {
	if len(parts) != 5 {
		abortTx("SET_SELL_TRIGGER needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return setSellTriggerCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (sst setSellTriggerCmd) Name() string {
	return fmt.Sprintf("[%d] SET_SELL_TRIGGER", sst.id)
}

func (sst setSellTriggerCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SET_SELL_TRIGGER</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, sst.id, sst.userID,
		sst.stock, sst.amount.ToFloat(),
	)

	return types.AuditEvent{
		UserID:    sst.userID,
		ID:        sst.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (sst setSellTriggerCmd) Execute() {
	autoTxKey := types.AutoTxKey{
		Stock:  sst.stock,
		UserID: sst.userID,
		Action: "Sell",
	}
	autoTx, found := workATXStore[autoTxKey]
	if !found {
		// autoTx not set. Fail this trans
		// TODO
	}

	autoTx.Trigger = sst.amount

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
			ContentType: "text/plain",
			Headers: amqp.Table{
				"transType": "autoTxInit",
			},
			Body: []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}
