package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

type commandStruct struct {
	Command string
}

var conn redis.Conn

func handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct{ Command string }{""}
	err := decoder.Decode(&cmd)

	failOnError(err, "Decoding Failed")

	//Validate Command
	//Add it to Redis
	_, err = conn.Do("RPUSH", pendingTxKey, cmd.Command)

	failOnError(err, "Failed to push to worker queue")
	responseText := fmt.Sprintf("Successfully performed %s\n", cmd.Command)
	fmt.Fprintf(w, responseText, r.URL.Path[1:])
}

func incomingTxWatcher() {
	conn = redisPool.Get()
	port := fmt.Sprintf(":%d", config.Redis.Port+*workerNum)
	consoleLog.Debugf("Started watching on port %s\n", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(port, nil)
}
