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

var reqCounter uint64

func handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct{ Command string }{""}
	err := decoder.Decode(&cmd)

	failOnError(err, "Decoding Failed")

	//Validate Command
	//Add it to Redis
	redisCmd := fmt.Sprintf("[%d] %s\n", reqCounter, cmd.Command)
	_, err = conn.Do("RPUSH", pendingTxKey, redisCmd)

	failOnError(err, "Failed to push to worker queue")
	reqCounter++
	responseText := fmt.Sprintf("Successfully performed %s\n", cmd.Command)
	fmt.Fprintf(w, responseText, r.URL.Path[1:])
}

func incomingTxWatcher() {

	conn = redisPool.Get()
	port := fmt.Sprintf(":%d", config.Redis.Port+*workerNum)
	fmt.Printf("Started watching on port %s\n", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(port, nil)
}
