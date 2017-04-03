package main

import (
	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

func rmqPush(ch *amqp.Channel, header string, body string) {
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
}

func sendAutoTxInit(autoTxInitChan <-chan types.AutoTxInit) {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	for {
		message := <-autoTxInitChan
		//TODO: check localquotecache. If we can fill it, don't even bother sending it.
		rmqPush(ch, "autoTxInit", message.ToCSV())
	}
}

func sendAutoTxCancel(autoTxCancelChan <-chan types.AutoTxKey) {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	for {
		message := <-autoTxCancelChan
		//TODO: check localquotecache. If we can fill it, don't even bother sending it.
		rmqPush(ch, "autoTxCancel", message.ToCSV())
	}
}
