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

	var cmd commandStruct
	err := decoder.Decode(&cmd)

	if err != nil {
		panic(err)
	}
	fmt.Println(cmd.Command)
	//Validate Command
	//Add it to Redis
	conn.Do("RPUSH", fmt.Sprintf("worker:%d:pendingtx", *workerNum), cmd.Command)

	responseText := fmt.Sprintf("Successfully performed %s\n", cmd.Command)
	fmt.Fprintf(w, responseText, r.URL.Path[1:])
}

func httpWatcher() {
	conn = redisPool.Get()
	port := fmt.Sprintf(":%d", 44431+*workerNum)
	fmt.Printf("Started watching on port %s\n", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(port, nil)
}
