package main

import (
	"math/rand"
	"time"

	types "github.com/distributeddesigns/shared_types"

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
			go cacheQuote(q)
		}
	}()

	<-done
}

func cacheQuote(q types.Quote) {
	quoteAge := time.Now().Unix() - q.Timestamp.Unix()
	ttl := config.QuotePolicy.BaseTTL - rand.Intn(config.QuotePolicy.BackoffTTL) - int(quoteAge)

	if ttl < config.QuotePolicy.MinTTL {
		consoleLog.Debugf("Not caching %s since TTL is %d", q.Stock, ttl)
		return
	}

	conn := redisPool.Get()
	defer conn.Close()

	quoteKey := getQuoteKey(q.Stock)
	serializedQuote := q.ToCSV()
	_, err := conn.Do("SETEX", quoteKey, ttl, serializedQuote)
	failOnError(err, "Could not update quote in redis")

	consoleLog.Debugf("Updated %s:%+v", quoteKey, serializedQuote)
	consoleLog.Debugf("%s will expire in %d sec", quoteKey, ttl)
}

func getQuoteKey(stock string) string {
	return redisBaseKey + "quotes:" + stock
}
