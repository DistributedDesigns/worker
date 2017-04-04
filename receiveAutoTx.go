package main

import (
	"fmt"
	"strconv"

	types "github.com/distributeddesigns/shared_types"
	"github.com/streadway/amqp"
)

const (
	aTxReceiveQueue = "autoTx_receive"
)

func receiveAutoTx() {
	// get channel
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Make sure all of the expected RabbitMQ exchanges and queues
	// exist before we start using them.
	// Recieve requests

	err = ch.ExchangeDeclare(
		autoTxExchange,      // name
		amqp.ExchangeDirect, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // args
	)
	failOnError(err, "Failed to declare exchange")
	queueName := fmt.Sprintf("%s:%d", aTxReceiveQueue, *workerNum)
	failOnError(err, "Failed to name queue")
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		strconv.Itoa(*workerNum), // routing key
		autoTxExchange,           // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to subscribe to autoTxExchange")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	failOnError(err, "Failed to consume on exchange")

	for d := range msgs {
		//fmt.Printf("Response Received\n")
		autoTxFilled, err := types.ParseAutoTxFilled(string(d.Body[:]))
		failOnError(err, "Failed to parse autoTxInit")
		consoleLog.Debugf("AutoTxFilled is : %+v\n", autoTxFilled)
		updateAccount(autoTxFilled)
	}
	fmt.Printf("AutoTxWorker Receiver Spinning\n")
}
