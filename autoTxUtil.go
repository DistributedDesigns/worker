package main

import (
	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

func updateAccount(autoTxFilled types.AutoTxFilled) {
	// TODO: Do account add and lock here, needs rebase
	return
}

func fulfillAutoTx(autoTxKey types.AutoTxKey, ch *amqp.Channel, header string, body string) {
	qr := types.QuoteRequest{
		Stock:      autoTxKey.Stock,
		UserID:     autoTxKey.UserID,
		AllowCache: true,
		ID:         ^uint64(0),
	}
	found := hasQuote(qr)
	if !found {
		// No quote found, let's autoTx that
		err := ch.Publish(
			"",          // exchange
			autoTxQueue, // routing key
			false,       // mandatory
			false,       // immediate
			amqp.Publishing{
				ContentType: "text/csv",
				Headers: amqp.Table{
					"transType": header,
				},
				Body: []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		return
	}
	// TODO: DO MATH FOR FILLED TRANS
	curr, err := currency.NewFromFloat(0.00)
	failOnError(err, "Failed to parse currency")
	autoTxFilled := types.AutoTxFilled{
		AddFunds:  curr,
		AddStocks: uint(0),
		AutoTxKey: autoTxKey,
	}
	updateAccount(autoTxFilled)
}

func sendAutoTxInit(autoTxInitChan <-chan types.AutoTxInit) {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	for {
		autoTxInit := <-autoTxInitChan
		autoTxKey := autoTxInit.AutoTxKey
		fulfillAutoTx(autoTxKey, ch, "autoTxInit", autoTxInit.ToCSV())
	}
}

func sendAutoTxCancel(autoTxCancelChan <-chan types.AutoTxKey) {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	for {
		autoTxKey := <-autoTxCancelChan
		fulfillAutoTx(autoTxKey, ch, "autoTxCancel", autoTxKey.ToCSV())
	}
}
