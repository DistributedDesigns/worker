package main

import "github.com/garyburd/redigo/redis"

func fetchNewTx(unprocessedTxs chan<- string) {
	conn := redisPool.Get()
	defer conn.Close()

	pendingTxKey := redisBaseKey + "pendingtx"

	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished watching", pendingTxKey)
			return
		default:
			// Block until something is returned from redis, or timeout
			r, err := redis.Values(conn.Do("BLPOP", pendingTxKey, pendingTxTimeout))
			if err == redis.ErrNil {
				consoleLog.Info("No new entries in", pendingTxKey)
				break
			} else if err != nil {
				failOnError(err, "Could not retrieve tx from redis")
			}

			// Convert reply to a string
			// BLPOP returns [key, val] pairs
			var key, tx string
			_, err = redis.Scan(r, &key, &tx)
			failOnError(err, "Could not convert transaction reply to string")

			// Blocks until there's room to insert the tx
			unprocessedTxs <- tx
		}
	}
}
