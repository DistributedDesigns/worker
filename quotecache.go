package main

import (
	"math/rand"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
	"github.com/garyburd/redigo/redis"

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
		redisBaseKey+":updater", // name
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

			consoleLog.Debugf(" [↙] Intercepted quote: %s %s", q.Stock, q.Price)
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
	return redisBaseKey + ":quotes:" + stock
}

// getQuote checks local redis for a quote
func getQuote(qr types.QuoteRequest) types.Quote {
	conn := redisPool.Get()
	defer conn.Close()

	quoteKey := getQuoteKey(qr.Stock)
	r, err := redis.String(conn.Do("GET", quoteKey))
	if err == redis.ErrNil {
		return getFreshQuote(qr)
	} else if err != nil {
		failOnError(err, "Could not retrieve quote from redis")
	}

	consoleLog.Debug(" [✔] Cache hit:", quoteKey)
	quote, err := types.ParseQuote(r)
	failOnError(err, "Could not parse quote from redis value")

	return quote
}

// getFreshQuote makes a request for a new quote to the quote service over RMQ
func getFreshQuote(qr types.QuoteRequest) types.Quote {
	consoleLog.Debug(" [x] Cache miss:", qr.ID, qr.Stock)

	freshQuotes := make(chan types.Quote, 1)
	ready := make(chan struct{}, 1)

	go watchForQuoteUpdate(qr, freshQuotes, ready)
	go requestQuote(qr, ready)

	return <-freshQuotes
}

func watchForQuoteUpdate(qr types.QuoteRequest, freshQuotes chan<- types.Quote, ready chan<- struct{}) {
	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Anonymous Q that filters for fresh stock broadcasts
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	var freshnessFilter string
	if qr.AllowCache {
		// Catch fresh and cached
		freshnessFilter = ".*"
	} else {
		freshnessFilter = ".fresh"
	}

	err = ch.QueueBind(
		q.Name, // name
		qr.Stock+freshnessFilter, // routing key
		quoteBroadcastEx,         // exchange
		false,                    // no-wait
		nil,                      // args
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		true,   // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Send the request for a new quote by unblocking requestQuote()
	consoleLog.Debug(" [-] Waiting for updates to", qr.Stock)
	ready <- struct{}{}

	// Hold here until RMQ quote update
	for d := range msgs {
		quote, err := types.ParseQuote(string(d.Body))
		failOnError(err, "Could not parse quote from RMQ")

		freshQuotes <- quote
		break
	}
}

func requestQuote(qr types.QuoteRequest, ready <-chan struct{}) {
	// Hold for quote watcher to create queue
	<-ready

	consoleLog.Debug(" [↑] Requesting new quote for", qr.Stock)

	ch, err := rmqConn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	header := amqp.Table{
		"serviceID": redisBaseKey,
	}

	err = ch.Publish(
		"",            // exchange
		quoteRequestQ, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			Headers:       header,
			CorrelationId: strconv.FormatInt(int64(qr.ID), 10),
			ContentType:   "text/plain",
			Body:          []byte(qr.ToCSV()),
		})
	failOnError(err, "Failed to publish a message")
}
