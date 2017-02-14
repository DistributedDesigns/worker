package main

import (
	"github.com/distributeddesigns/shared_types"

	"github.com/streadway/amqp"
)

func initQuoteCacheRMQ() {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Send quote requests
	_, err = ch.QueueDeclare(
		quoteRequestQ, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Catch quote updates
	err = ch.ExchangeDeclare(
		quoteBroadcastEx,   // name
		amqp.ExchangeTopic, // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // args
	)
	failOnError(err, "Failed to declare an exchange")
}

func catchQuoteBroadcasts() {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		redisBaseKey+"updater", // name
		true,  // durable
		true,  // delete when unused
		false, // exclusive
		false, // no wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,           // name
		"#",              // routing key
		quoteBroadcastEx, // exchange
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		consoleLog.Info(" [-] Watching for quote updates on", quoteBroadcastEx)

		for d := range msgs {
			q, err := types.ParseQuote(string(d.Body))
			if err != nil {
				consoleLog.Errorf("Caught a bad quote: %s, %s", string(d.Body), err)
				break
			}

			consoleLog.Infof(" [↙] Intercepted quote: %s %s", q.Stock, q.Price)
			cacheQuote(q)
		}
	}()

	<-done
}

func cacheQuote(q types.Quote) {
	consoleLog.Warning("You really should cache this:", q.Stock)
}
