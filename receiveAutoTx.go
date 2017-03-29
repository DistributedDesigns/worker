package main

import (
	"fmt"
	"strconv"
)

func receiveAutoTx() {
	// get channel
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	// //consume on
	// defer ch.Close()
	defer ch.Close()

	// Make sure all of the expected RabbitMQ exchanges and queues
	// exist before we start using them.
	// Recieve requests

	failOnError(err, "Failed to declare exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no wait
		nil,   // arguments
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
		fmt.Printf("Response Received\n")
		fmt.Println(d)
	}

	// failOnError(err, "Failed to consume from quoteBroadcast Channel")
	fmt.Printf("AutoTxWorker Receiver Spinning\n")
}
