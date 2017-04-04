package main

import (
	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

func updateAccount(autoTxFilled types.AutoTxFilled) {
	accountStore[autoTxFilled.AutoTxKey.UserID].Lock()
	accountStore[autoTxFilled.AutoTxKey.UserID].AddFunds(autoTxFilled.AddFunds)
	accountStore[autoTxFilled.AutoTxKey.UserID].AddStock(autoTxFilled.AutoTxKey.Stock, uint64(autoTxFilled.AddStocks))
	accountStore[autoTxFilled.AutoTxKey.UserID].Unlock()
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
				var filledStock uint
				var filledCash currency.Currency
				if autoTxKey.Action == "Buy" {
					filledStock, filledCash = quote.Price.FitsInto(autoTxInit.Amount)
					failOnError(err, "Failed to parse currency")
				} else {
					numStock, remCash := autoTxInit.Trigger.FitsInto(autoTxInit.Amount) // amount of stock we reserved from their port
					filledCash = quote.Price
					err = filledCash.Mul(float64(numStock))
					filledCash.Add(remCash) // Re-add the unfilled value
					filledStock = 0
					failOnError(err, "Failed to parse currency")
				}

				autoTxFilled := types.AutoTxFilled{
					AddFunds:  filledCash,
					AddStocks: filledStock,
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
