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

func sendAutoTx() {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	for {
		select {
		case autoTxInit := <-autoTxInitChan:
			autoTxKey := autoTxInit.AutoTxKey

			qr := types.QuoteRequest{
				Stock:  autoTxKey.Stock,
				UserID: autoTxKey.UserID,
			}
			quote, found := getCachedQuote(qr)
			var validTrigger bool
			if autoTxKey.Action == "Buy" {
				validTrigger = found && (quote.Price.ToFloat() < autoTxInit.Trigger.ToFloat())
			} else {
				validTrigger = found && (quote.Price.ToFloat() > autoTxInit.Trigger.ToFloat())
			}

			if validTrigger {
				// TODO: DO MATH FOR FILLED TRANS
				curr, err := currency.NewFromFloat(0.00)
				failOnError(err, "Failed to parse currency")
				autoTxFilled := types.AutoTxFilled{
					AddFunds:  curr,
					AddStocks: uint(0),
					AutoTxKey: autoTxKey,
				}
				updateAccount(autoTxFilled)
				continue
			}
			// No quote found, let's autoTx that
			err := ch.Publish(
				"",          // exchange
				autoTxQueue, // routing key
				false,       // mandatory
				false,       // immediate
				amqp.Publishing{
					ContentType: "text/csv",
					Headers: amqp.Table{
						"transType": "autoTxInit",
					},
					Body: []byte(autoTxInit.ToCSV()),
				})
			failOnError(err, "Failed to publish a message")

		case autoTxKey := <-autoTxCancelChan:
			err := ch.Publish(
				"",          // exchange
				autoTxQueue, // routing key
				false,       // mandatory
				false,       // immediate
				amqp.Publishing{
					ContentType: "text/csv",
					Headers: amqp.Table{
						"transType": "autoTxCancel",
					},
					Body: []byte(autoTxKey.ToCSV()),
				})
			failOnError(err, "Failed to publish a message")
		}

	}
}
